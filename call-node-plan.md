# Implementation Plan: CALL Node — Cross-Flow Invocation

## Context

The design has been ratified in GitHub Discussion [#2639](https://github.com/thunder-id/thunderid/discussions/2639). We are now implementing it.

The flow engine currently executes a single self-contained graph per `flow/execute` invocation. This plan introduces a new node type `CALL` that transfers execution to a referenced flow and returns control to the caller when the callee's `END` is reached. The call/return mechanism uses a per-frame stack of execution state while keeping identity-bearing data (user inputs, authenticated user, history, application, assertion) shared across frames.

Decisions in force (from the design discussion + user clarifications during planning):

- **No restrictions on what a flow can call.** Any flow can call any other flow. The CALL node carries a flow ID in `ref`; the engine resolves it via `flowMgtService.GetGraph(ref)` and runs whatever comes back. There is no flow-type restriction and no app-attachment check at the engine layer.
- **EngineContext refactor is narrow.** Only the call-node-touched fields adopt the new shared-vs-frame split; existing public fields stay as-is. The frame stack is internal; executors keep reading `ctx.Graph` / `ctx.CurrentNode` / `ctx.RuntimeData` etc. — these reflect the active frame because the engine swaps them on push/pop.
- **Error propagation.** Callee `END`-with-failure carries the failure error in `NodeResponse.Error`; engine pops the frame and forwards to caller's `onFailure`. The failure response is consumed by the next prompt and then cleared from context so it does not leak into subsequent caller execution.
- **Test coverage target.** 100% (or max-possible) patch coverage for every changed file — backend Go and frontend TypeScript.
- **Integration tests are end-to-end against a running server** and live in [tests/integration/flow/](tests/integration/flow/) (one of `authentication/`, `recovery/`, `registration/`). Backend Go unit tests still cover the engine mechanics in isolation, but the three CALL scenarios called out below are real HTTP-driven integration tests.

## Critical Files

### Backend — new files

- [backend/internal/flow/core/call_node.go](backend/internal/flow/core/call_node.go) — `CallNodeInterface`, `callNode` impl, factory constructor.
- [backend/internal/flow/core/call_node_test.go](backend/internal/flow/core/call_node_test.go) — unit tests.

### Backend — files to modify

- [backend/internal/flow/common/constants.go](backend/internal/flow/common/constants.go) — add `NodeTypeCall NodeType = "CALL"`; add `NodeStatusCall NodeStatus = "CALL_FLOW"`; add error constants for unresolved CALL target / depth exceeded.
- [backend/internal/flow/common/model.go](backend/internal/flow/common/model.go) — add `CallTargetFlowID string` field on `NodeResponse` (set by CALL node's `Execute`, consumed by the engine in post-processing). Also extend the `RICH_TEXT` prompt component metadata with an optional action wiring (see "Schema additions" below).
- [backend/internal/flow/core/factory.go](backend/internal/flow/core/factory.go) — add `case common.NodeTypeCall` in `CreateNode` switch (around line 60); extend `CloneNode` to copy `CallNodeInterface` fields (mirrors the lines 134-163 pattern for representation / executor / prompt nodes).
- [backend/internal/flow/mgt/model.go](backend/internal/flow/mgt/model.go) — add `Flow *FlowReferenceDefinition` to `NodeDefinition` with `json:"flow"`; define `FlowReferenceDefinition { Ref string }`.
- [backend/internal/flow/mgt/graph_builder.go](backend/internal/flow/mgt/graph_builder.go):
  - `processNode`: extend `isFinalNode` to consider CALL edges; add `configureCallNodeReference(nodeDef, node)`.
  - `configureNodeNavigation`: when node type is `CALL`, set `onSuccess` (required) and `onFailure` (optional) on the `CallNodeInterface`; reject `onIncomplete` for CALL.
  - Add `validateCallNodeDefinition`: enforce non-empty `flow.ref`, presence of `onSuccess`, absence of `onIncomplete`, no `prompts` / `executor` blocks.
- [backend/internal/flow/flowexec/model.go](backend/internal/flow/flowexec/model.go):
  - Private `frame` struct + accessors (`graph`, `flowType`, `currentNode`, `currentNodeResponse`, `currentAction`, `currentSegmentID`, `runtimeData`, `forwardedData`, `additionalData`, `resumeCallNodeID`).
  - `frameStack []*frame` (private) on `EngineContext` + `PushFrame()`, `PopFrame()`, `FrameDepth()`, `TopFrame()` methods.
  - `SharedRuntimeData map[string]string` (the persisted shared bucket) + `RuntimeData(key string) (string, bool)` (frame-local first, then shared) and `SetRuntimeData(key, value string, shared bool)`.
  - Extend `flowContextContent` with `FrameStack *string` (JSON-serialized list of `{ graphId, currentNodeId, currentAction, currentSegmentId, runtimeData, forwardedData, additionalData, resumeCallNodeId }`) and `SharedRuntimeData *string`. Round-trip in `ToEngineContext` / `FromEngineContext`.
- [backend/internal/flow/flowexec/engine.go](backend/internal/flow/flowexec/engine.go):
  - `processNodeResponse`: add case `common.NodeStatusCall` → `handleCallResponse` (push frame, swap context to callee, return callee's start node).
  - `handleCompletedResponse`: when current node is `END` and `len(frameStack) > 0`, pop, restore frame, return caller CALL's `onSuccess` node.
  - Failure path in `processNodeResponse`: when stack is non-empty, do not terminate the whole flow; pop, restore frame, route to caller CALL's `onFailure` (synthesizing a response carrying the failure error so the next prompt receives it; clear from `ctx` after the next prompt consumes it). If `onFailure` absent, terminate with the same error.
  - Hard-coded `maxCallDepth = 5` package const; checked in `handleCallResponse` before push.
  - Resolve the callee graph via `flowMgtService.GetGraph(callTargetFlowID)`. Engine gains a `flowGraphResolver` dependency injected at construction (parallel to `executorRegistry`).
- [backend/internal/flow/flowexec/service.go](backend/internal/flow/flowexec/service.go):
  - Implement the `flowGraphResolver` (thin wrapper around `flowMgtService.GetGraph`).
  - Persistence: the active frame's `GraphID` is what `flowContextContent.GraphID` stores, so the existing `getFlowGraph` resume path keeps working unchanged.

### Backend — unit tests (target 100% / max patch coverage)

- [backend/internal/flow/core/call_node_test.go](backend/internal/flow/core/call_node_test.go) — factory, getters/setters, `Execute` response shape (returns `NodeStatusCall` with `CallTargetFlowID = referencedFlow`).
- [backend/internal/flow/core/factory_test.go](backend/internal/flow/core/factory_test.go) — extend with CALL `CreateNode` and `CloneNode` cases.
- [backend/internal/flow/mgt/graph_builder_test.go](backend/internal/flow/mgt/graph_builder_test.go) — CALL node parsing (valid), missing `flow.ref`, `onIncomplete` rejected, `prompts` / `executor` rejected on CALL.
- [backend/internal/flow/flowexec/engine_test.go](backend/internal/flow/flowexec/engine_test.go) — push/pop semantics with mocked graphs: simple call+return, nested (depth 2), depth limit exceeded, suspend in callee + resume, callee failure with caller `onFailure`, callee failure without caller `onFailure`, unresolved `ref`.
- [backend/internal/flow/flowexec/model_test.go](backend/internal/flow/flowexec/model_test.go) — round-trip persistence with non-empty frame stack and non-empty shared runtime data.

### Backend — integration tests

Real HTTP integration tests against a running server, in [tests/integration/flow/](tests/integration/flow/), following the existing pattern (e.g. [basic_auth_test.go](tests/integration/flow/authentication/basic_auth_test.go)). Each test provisions an application + flows via API, exercises the flow through `flow/execute`, and asserts the final response. Each scenario covers both straight-through completion and a one-step suspend/resume.

- [tests/integration/flow/authentication/call_to_registration_test.go](tests/integration/flow/authentication/call_to_registration_test.go) — authentication flow whose credentials prompt offers a "register" action that routes to a CALL referencing a registration flow; on registration completion the caller's `auth_assert` runs and the flow completes with an assertion. Also covers a callee-failure path through the caller's `onFailure`.
- [tests/integration/flow/authentication/call_to_recovery_test.go](tests/integration/flow/authentication/call_to_recovery_test.go) — authentication flow that calls a recovery flow (e.g. after failure / on "forgot password"); on recovery completion the caller resumes (auto-login via `auth_assert`).
- [tests/integration/flow/registration/call_to_authentication_test.go](tests/integration/flow/registration/call_to_authentication_test.go) — registration flow that, after collecting an identifier already known to the system, calls into an authentication flow to log the existing user in.

## Schema additions (backend)

`RICH_TEXT` prompt-component metadata gains an optional `action` field with the same shape as the `ACTION` component's action wiring (`ref`, `eventType`). When set, the component is interactive (clickable, drives a flow action). When absent, the component remains pure display (current behavior). The graph builder validates `action.ref` exists in the prompt's action set if specified. This is what lets the credentials prompt's "Sign up" rich text fire a flow action rather than navigating to an external URL.

## Frontend — SDK and Gate

The renderer for prompt components lives in the SDK and is consumed by the Gate. The rich-text → action wiring is therefore a primarily SDK change, with a small Gate change only if the Gate overrides the SDK's rich-text rendering (we will inspect both during implementation).

- [sdks/javascript/src/models/v2/embedded-flow-v2.ts](sdks/javascript/src/models/v2/embedded-flow-v2.ts) — extend the `RICH_TEXT` model type with the optional `action` field so the rich-text renderer can read it.
- [sdks/react/src/components/presentation/auth/AuthOptionFactory.tsx](sdks/react/src/components/presentation/auth/AuthOptionFactory.tsx) — where the rich-text class is currently applied; update the rich-text renderer to:
  - look for sentinel-marked anchors inside the rich-text HTML (e.g. `data-action-ref="..."`) when the component carries an `action`,
  - bind click on those anchors to dispatch a `flow/execute` submission with the action ref instead of navigating,
  - leave pure-display rich text behaving exactly as today.
- [frontend/apps/gate/](frontend/apps/gate/) — inspect for any local override of the SDK rich-text renderer (e.g. in `SignIn` / `SignInBox`); if Gate overrides, apply the same action-binding change there.

### Frontend — Flow Builder (Console)

[frontend/apps/console/src/features/flows/](frontend/apps/console/src/features/flows/) — the Flow Builder lives here (it is part of the Console app, not a standalone composer; see [elements.ts](frontend/apps/console/src/features/flows/models/elements.ts), [generateFlowGraph.ts](frontend/apps/console/src/features/flows/utils/generateFlowGraph.ts), and [resource-property-panel/rich-text/](frontend/apps/console/src/features/flows/components/resource-property-panel/rich-text/)):

- **CALL node**: add to the node-type palette and canvas renderer (distinct icon, label "Call flow"). Configuration sidebar fields: `flow.ref` (dropdown sourced from **all existing flows** in the system, minus the flow currently being edited); `onSuccess` (required); `onFailure` (optional).
- **"Open referenced flow" affordance** on a selected CALL node: navigates the composer to the referenced flow. Guard against losing in-progress unsaved edits — show an "Unsaved changes — save, discard, or cancel?" confirmation before navigating. Reuse the existing unsaved-changes guard if one exists; add one if not.
- **Rich-text action configuration**: the rich-text component editor in `resource-property-panel/rich-text/` must let the author wire an `action` (the same `action` field added to the backend `RICH_TEXT` schema) to a sentinel-marked anchor inside the rich-text content. This makes the rich-text-action wiring authorable from the builder UI, not just hand-coded in JSON.
- **Existing "Sign up" widget**: the Flow Builder ships with pre-canned widgets that drop a "Sign up" rich-text snippet into a prompt node. Update this widget so it now also adds an `action` (with a default ref like `action_signup`) into the prompt node's action set, and wires the rich-text anchor to that action. Without this, authors using the widget would still produce a pure-display rich text.

### Frontend — unit tests

Target 100% / max-possible patch coverage on every changed `.ts` / `.tsx`:

- SDK rich-text renderer: tests for action-bearing rich text (click triggers action dispatch) and pure-display rich text (click navigates as before).
- SDK model: tests for the new optional `action` field on `RICH_TEXT`.
- Gate: if Gate overrides the SDK renderer, the equivalent renderer-level tests.
- Flow Builder: tests for CALL palette entry, CALL canvas rendering, sidebar config (ref dropdown population, edge fields), unsaved-changes guard before "Open referenced flow", rich-text editor action configuration, updated Sign-up widget producing the action-bearing rich text.

## Implementation Order

Each stage is a self-contained PR (single commit per PR per [AGENTS.md](AGENTS.md)).

### Stage 1 — Backend node + schema + graph builder

1. Constants in [common/constants.go](backend/internal/flow/common/constants.go); `CallTargetFlowID` in [common/model.go](backend/internal/flow/common/model.go); `RICH_TEXT` action field in the component metadata.
2. [core/call_node.go](backend/internal/flow/core/call_node.go) with `CallNodeInterface` (extending `NodeInterface`, with `GetReferencedFlow / SetReferencedFlow`, `GetOnSuccess / SetOnSuccess`, `GetOnFailure / SetOnFailure`) and `callNode` impl whose `Execute()` returns `&NodeResponse{Status: NodeStatusCall, CallTargetFlowID: n.referencedFlow}`. Mirror [representation_node.go](backend/internal/flow/core/representation_node.go).
3. Register CALL in [core/factory.go](backend/internal/flow/core/factory.go) `CreateNode` switch and extend `CloneNode` to copy CALL fields.
4. Extend [mgt/model.go](backend/internal/flow/mgt/model.go) (`Flow` field, `FlowReferenceDefinition`).
5. Extend [mgt/graph_builder.go](backend/internal/flow/mgt/graph_builder.go) to parse the `flow` block, set CALL edges, reject `onIncomplete` / `prompts` / `executor` on CALL, require `onSuccess` and `flow.ref`. Validate the new `RICH_TEXT.action.ref` references a configured action.
6. Unit tests for the node, factory, schema, and graph builder.

After this stage: CALL nodes can be defined in flow JSON and built into graphs; the engine doesn't execute them yet (returns an error on the new status).

### Stage 2 — Backend engine frame stack, call/return, persistence

1. Frame stack + shared-vs-frame split in [flowexec/model.go](backend/internal/flow/flowexec/model.go). Extend `flowContextContent`, `ToEngineContext`, `FromEngineContext`.
2. `flowGraphResolver` interface implemented in [flowexec/service.go](backend/internal/flow/flowexec/service.go) (one method that wraps `flowMgtService.GetGraph`). Injected into the engine via `newFlowEngine`.
3. In [flowexec/engine.go](backend/internal/flow/flowexec/engine.go):
   - `maxCallDepth = 5` package const.
   - `handleCallResponse`: validate depth, resolve target, snapshot current frame (record current CALL node id as `resumeCallNodeID`), push, swap `ctx.Graph` / `ctx.FlowType` / `ctx.CurrentNode` / `ctx.CurrentSegmentID` / `ctx.RuntimeData` / `ctx.ForwardedData` / `ctx.AdditionalData` to callee start state, return callee's start node.
   - Adapt `handleCompletedResponse` so that on `END` with non-empty stack: pop, restore frame, return caller CALL's `onSuccess`.
   - Adapt failure path: when stack non-empty, pop and route to caller CALL's `onFailure` (synthesizing a response carrying the failure error). Clear failure from `ctx` after the next prompt consumes it. If `onFailure` absent, terminate.
   - Persist frame stack on suspend; rehydrate on resume.
4. Engine unit tests covering all paths.

### Stage 3 — Integration tests

1. [tests/integration/flow/authentication/call_to_registration_test.go](tests/integration/flow/authentication/call_to_registration_test.go).
2. [tests/integration/flow/authentication/call_to_recovery_test.go](tests/integration/flow/authentication/call_to_recovery_test.go).
3. [tests/integration/flow/registration/call_to_authentication_test.go](tests/integration/flow/registration/call_to_authentication_test.go).

Each provisions the required application + flow definitions over the management API, runs the journey through `flow/execute`, and asserts the final assertion / status. Each test covers straight-through completion and a one-step suspend/resume; the auth→registration test additionally covers callee failure with caller `onFailure`.

### Stage 4 — SDK + Gate rich-text action

1. Extend [sdks/javascript/src/models/v2/embedded-flow-v2.ts](sdks/javascript/src/models/v2/embedded-flow-v2.ts) `RICH_TEXT` with the optional `action` field.
2. Update the rich-text renderer in [sdks/react/src/components/presentation/auth/AuthOptionFactory.tsx](sdks/react/src/components/presentation/auth/AuthOptionFactory.tsx) to bind sentinel-marked anchors to a `flow/execute` action dispatch when `action` is set.
3. Inspect [frontend/apps/gate/](frontend/apps/gate/) for any local override of the SDK renderer; apply the same change if so.
4. Unit tests for both pure-display and action-bearing rich-text paths.

### Stage 5 — Flow Builder UI

1. CALL palette entry + canvas rendering in [frontend/apps/console/src/features/flows/](frontend/apps/console/src/features/flows/).
2. Configuration sidebar with `ref` dropdown (all flows minus self), `onSuccess`, `onFailure`.
3. "Open referenced flow" affordance with unsaved-changes confirmation guard.
4. Rich-text editor: add UI for configuring the optional `action` field on a rich-text component; sentinel anchor insertion.
5. Update the pre-canned "Sign up" widget to emit a rich-text component with an `action` and the matching prompt action.
6. Unit tests for all of the above.

## Persistence Schema Notes

- `flow_context.CONTEXT` JSON gains two fields:
  - `frameStack`: JSON-encoded array of `{ graphId, currentNodeId, currentAction, currentSegmentId, runtimeData, forwardedData, additionalData, resumeCallNodeId }`. Empty / absent for flows that have not crossed into a callee.
  - `sharedRuntimeData`: JSON-encoded string-string map. Absent / `{}` for backward compatibility — existing executors continue to write to `RuntimeData` (frame-local) and read transparently.
- No DB migration required; only the JSON blob inside the existing `CONTEXT` column changes.
- The active frame's `GraphID` is what `flowContextContent.GraphID` stores so the existing `getFlowGraph` resume path works unchanged.

## Verification

- `make lint` and `make test` from the repo root (Go + frontend). Patch coverage gates must hit 100% / max-possible on every file touched.
- The three integration tests under [tests/integration/flow/](tests/integration/flow/) pass against the live test server.
- Negative cases: 6-deep CALL chain returns `ErrorCallDepthExceeded` at the 6th call; a CALL whose `ref` resolves to no flow returns a clean engine error rather than a panic.
- Manual end-to-end (after Stages 4-5 ship): in a test deployment, build a flow in the Console UI that uses a CALL into a registration flow and a rich-text "Sign up" action on the credentials prompt; complete the journey through the Gate; observe the assertion delivered by the caller's `auth_assert`. Edit the flow, click "Open referenced flow" while there are unsaved changes — confirm the guard fires.
