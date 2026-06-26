/*
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package flowexec

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	appmodel "github.com/thunder-id/thunderid/internal/application/model"
	authncm "github.com/thunder-id/thunderid/internal/authn/common"
	authnprovidercm "github.com/thunder-id/thunderid/internal/authnprovider/common"
	managerpkg "github.com/thunder-id/thunderid/internal/authnprovider/manager"
	"github.com/thunder-id/thunderid/internal/flow/common"
	"github.com/thunder-id/thunderid/internal/flow/core"
	"github.com/thunder-id/thunderid/internal/system/error/apierror"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
)

// frame captures the per-call execution state that must be saved when a CALL node pushes
// execution into a callee flow and restored when the callee returns.
type frame struct {
	graph               core.GraphInterface
	flowType            common.FlowType
	currentNode         core.NodeInterface
	currentNodeResponse *common.NodeResponse
	currentAction       string
	currentSegmentID    string
	runtimeData         map[string]string
	forwardedData       map[string]interface{}
	additionalData      map[string]string
	// resumeCallNodeID is the ID of the CALL node in the caller graph that triggered this frame.
	// On pop the engine uses it to look up onSuccess / onFailure.
	resumeCallNodeID string
}

// EngineContext holds the overall context used by the flow engine during execution.
type EngineContext struct {
	Context context.Context

	ExecutionID    string
	FlowType       common.FlowType
	AppID          string
	Verbose        bool
	UserInputs     map[string]string
	RuntimeData    map[string]string
	ForwardedData  map[string]interface{}
	AdditionalData map[string]string
	TraceID        string

	CurrentNode         core.NodeInterface
	CurrentNodeResponse *common.NodeResponse
	CurrentAction       string
	CurrentSegmentID    string

	Graph       core.GraphInterface
	Application appmodel.Application

	AuthenticatedUser authncm.AuthenticatedUser
	AuthUser          managerpkg.AuthUser
	Assertion         string
	ExecutionHistory  map[string]*common.NodeExecutionRecord

	ChallengeTokenIn   string
	ChallengeTokenHash string

	// frameStack holds the saved call frames. The top of the stack is the most recent caller.
	frameStack []*frame
	// SharedRuntimeData is a cross-frame key-value store written by
	// SetSharedRuntimeData and read by GetRuntimeData when a key is absent from the active frame.
	SharedRuntimeData map[string]string
}

// PushFrame saves the current execution state as a new frame and returns it.
// The caller is responsible for updating EngineContext fields to the callee state after pushing.
func (e *EngineContext) PushFrame(resumeCallNodeID string) {
	f := &frame{
		graph:               e.Graph,
		flowType:            e.FlowType,
		currentNode:         e.CurrentNode,
		currentNodeResponse: e.CurrentNodeResponse,
		currentAction:       e.CurrentAction,
		currentSegmentID:    e.CurrentSegmentID,
		runtimeData:         e.RuntimeData,
		forwardedData:       e.ForwardedData,
		additionalData:      e.AdditionalData,
		resumeCallNodeID:    resumeCallNodeID,
	}
	e.frameStack = append(e.frameStack, f)
}

// PopFrame restores the most-recently-pushed frame and removes it from the stack.
// Returns nil when the stack is empty (caller must check).
func (e *EngineContext) PopFrame() *frame {
	if len(e.frameStack) == 0 {
		return nil
	}
	top := e.frameStack[len(e.frameStack)-1]
	e.frameStack = e.frameStack[:len(e.frameStack)-1]
	e.Graph = top.graph
	e.FlowType = top.flowType
	e.CurrentNode = top.currentNode
	e.CurrentNodeResponse = top.currentNodeResponse
	e.CurrentAction = top.currentAction
	e.CurrentSegmentID = top.currentSegmentID
	e.RuntimeData = top.runtimeData
	e.ForwardedData = top.forwardedData
	e.AdditionalData = top.additionalData
	return top
}

// FrameDepth returns the number of saved frames (0 means we are in the root flow).
func (e *EngineContext) FrameDepth() int {
	return len(e.frameStack)
}

// TopFrame returns the topmost frame without removing it, or nil if the stack is empty.
func (e *EngineContext) TopFrame() *frame {
	if len(e.frameStack) == 0 {
		return nil
	}
	return e.frameStack[len(e.frameStack)-1]
}

// SetSharedRuntimeData writes a value into the cross-frame shared runtime data bucket.
func (e *EngineContext) SetSharedRuntimeData(key, value string) {
	if e.SharedRuntimeData == nil {
		e.SharedRuntimeData = make(map[string]string)
	}
	e.SharedRuntimeData[key] = value
}

// GetSharedRuntimeData returns a value from the cross-frame shared runtime data bucket.
func (e *EngineContext) GetSharedRuntimeData(key string) (string, bool) {
	if e.SharedRuntimeData == nil {
		return "", false
	}
	v, ok := e.SharedRuntimeData[key]
	return v, ok
}

// FlowStep represents the outcome of a individual flow step
type FlowStep struct {
	ExecutionID    string
	StepID         string
	Type           common.FlowStepType
	Status         common.FlowStatus
	ChallengeToken string
	Data           FlowData
	Assertion      string
	Error          *serviceerror.ServiceError
}

// FlowData holds the data returned by a flow execution step
type FlowData struct {
	Inputs         []common.Input      `json:"inputs,omitempty"`
	RedirectURL    string              `json:"redirectURL,omitempty"`
	Actions        []common.Action     `json:"actions,omitempty"`
	Meta           interface{}         `json:"meta,omitempty"`
	AdditionalData map[string]string   `json:"additionalData,omitempty"`
	FieldErrors    []common.FieldError `json:"fieldErrors,omitempty"`
}

// FlowResponse represents the flow execution API response body
type FlowResponse struct {
	ExecutionID    string                  `json:"executionId"`
	StepID         string                  `json:"stepId,omitempty"`
	FlowStatus     string                  `json:"flowStatus"`
	Type           string                  `json:"type,omitempty"`
	ChallengeToken string                  `json:"challengeToken,omitempty"`
	Data           FlowData                `json:"data,omitempty"`
	Assertion      string                  `json:"assertion,omitempty"`
	Error          *apierror.ErrorResponse `json:"error,omitempty"`
}

// FlowRequest represents the flow execution API request body
type FlowRequest struct {
	ApplicationID  string            `json:"applicationId"`
	FlowType       string            `json:"flowType"`
	Verbose        bool              `json:"verbose,omitempty"`
	ExecutionID    string            `json:"executionId"`
	ChallengeToken string            `json:"challengeToken,omitempty"`
	Action         string            `json:"action"`
	Inputs         map[string]string `json:"inputs"`
}

// FlowInitContext represents the context for initiating a new flow with runtime data
type FlowInitContext struct {
	ApplicationID string
	FlowType      string
	RuntimeData   map[string]string
	InitialInputs map[string]string
	ExpirySeconds int64
}

// FlowContextDB represents the database row for a flow context.
type FlowContextDB struct {
	ExecutionID string
	Context     string
	ExpiryTime  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// serializedFrame is the on-disk representation of a single call frame.
type serializedFrame struct {
	GraphID          string  `json:"graphId"`
	CurrentNodeID    *string `json:"currentNodeId,omitempty"`
	CurrentAction    *string `json:"currentAction,omitempty"`
	CurrentSegmentID *string `json:"currentSegmentId,omitempty"`
	RuntimeData      *string `json:"runtimeData,omitempty"`
	ForwardedData    *string `json:"forwardedData,omitempty"`
	AdditionalData   *string `json:"additionalData,omitempty"`
	ResumeCallNodeID string  `json:"resumeCallNodeId,omitempty"`
}

// flowContextContent holds all flow state serialized into the CONTEXT JSON column.
type flowContextContent struct {
	AppID               string  `json:"appId"`
	Verbose             bool    `json:"verbose"`
	CurrentNodeID       *string `json:"currentNodeId,omitempty"`
	CurrentAction       *string `json:"currentAction,omitempty"`
	CurrentSegmentID    *string `json:"currentSegmentId,omitempty"`
	GraphID             string  `json:"graphId"`
	RuntimeData         *string `json:"runtimeData,omitempty"`
	ExecutionHistory    *string `json:"executionHistory,omitempty"`
	IsAuthenticated     bool    `json:"isAuthenticated"`
	UserID              *string `json:"userId,omitempty"`
	OUID                *string `json:"ouId,omitempty"`
	UserType            *string `json:"userType,omitempty"`
	UserInputs          *string `json:"userInputs,omitempty"`
	UserAttributes      *string `json:"userAttributes,omitempty"`
	Token               *string `json:"token,omitempty"`
	AvailableAttributes *string `json:"availableAttributes,omitempty"`
	AuthUser            *string `json:"authUser,omitempty"`
	ChallengeTokenHash  *string `json:"challengeTokenHash,omitempty"`
	FrameStack          *string `json:"frameStack,omitempty"`
	SharedRuntimeData   *string `json:"sharedRuntimeData,omitempty"`
}

// GraphResolverFunc resolves a flow graph by its ID.
// Used during context deserialization to hydrate saved call frames.
type GraphResolverFunc func(ctx context.Context, graphID string) (core.GraphInterface, error)

// GetGraphID extracts the graph ID from the context JSON.
func (f *FlowContextDB) GetGraphID(_ context.Context) (string, error) {
	var content flowContextContent
	if err := json.Unmarshal([]byte(f.Context), &content); err != nil {
		return "", err
	}
	return content.GraphID, nil
}

// deserializeFrameStack reconstructs the saved call frames from the persisted content.
// Returns nil without error when FrameStack is absent or resolveGraph is nil.
func deserializeFrameStack(ctx context.Context, content flowContextContent,
	resolveGraph GraphResolverFunc) ([]*frame, error) {
	if content.FrameStack == nil || resolveGraph == nil {
		return nil, nil
	}
	var serializedFrames []serializedFrame
	if err := json.Unmarshal([]byte(*content.FrameStack), &serializedFrames); err != nil {
		return nil, err
	}
	frames := make([]*frame, 0, len(serializedFrames))
	for _, sf := range serializedFrames {
		frameGraph, err := resolveGraph(ctx, sf.GraphID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve frame graph %s: %w", sf.GraphID, err)
		}
		var currentNode core.NodeInterface
		if sf.CurrentNodeID != nil {
			if n, exists := frameGraph.GetNode(*sf.CurrentNodeID); exists {
				currentNode = n
			}
		}
		var currentAction, currentSegmentID string
		if sf.CurrentAction != nil {
			currentAction = *sf.CurrentAction
		}
		if sf.CurrentSegmentID != nil {
			currentSegmentID = *sf.CurrentSegmentID
		}
		var runtimeData map[string]string
		if sf.RuntimeData != nil {
			if err := json.Unmarshal([]byte(*sf.RuntimeData), &runtimeData); err != nil {
				return nil, err
			}
		}
		var forwardedData map[string]interface{}
		if sf.ForwardedData != nil {
			if err := json.Unmarshal([]byte(*sf.ForwardedData), &forwardedData); err != nil {
				return nil, err
			}
		}
		var additionalData map[string]string
		if sf.AdditionalData != nil {
			if err := json.Unmarshal([]byte(*sf.AdditionalData), &additionalData); err != nil {
				return nil, err
			}
		}
		frames = append(frames, &frame{
			graph:            frameGraph,
			flowType:         frameGraph.GetType(),
			currentNode:      currentNode,
			currentAction:    currentAction,
			currentSegmentID: currentSegmentID,
			runtimeData:      runtimeData,
			forwardedData:    forwardedData,
			additionalData:   additionalData,
			resumeCallNodeID: sf.ResumeCallNodeID,
		})
	}
	return frames, nil
}

// serializeFrameStack converts the in-memory call-frame stack into a JSON string pointer
// suitable for storage. Returns nil when the stack is empty.
func serializeFrameStack(frameStack []*frame) (*string, error) {
	if len(frameStack) == 0 {
		return nil, nil
	}
	serializedFrames := make([]serializedFrame, 0, len(frameStack))
	for _, f := range frameStack {
		if f.graph == nil || f.graph.GetID() == "" {
			return nil, fmt.Errorf("frame graph with a valid ID is required to persist frame stack")
		}
		sf := serializedFrame{
			GraphID:          f.graph.GetID(),
			ResumeCallNodeID: f.resumeCallNodeID,
		}
		if f.currentNode != nil {
			nodeID := f.currentNode.GetID()
			sf.CurrentNodeID = &nodeID
		}
		if f.currentAction != "" {
			sf.CurrentAction = &f.currentAction
		}
		if f.currentSegmentID != "" {
			sf.CurrentSegmentID = &f.currentSegmentID
		}
		if len(f.runtimeData) > 0 {
			b, err := json.Marshal(f.runtimeData)
			if err != nil {
				return nil, err
			}
			s := string(b)
			sf.RuntimeData = &s
		}
		if len(f.forwardedData) > 0 {
			b, err := json.Marshal(f.forwardedData)
			if err != nil {
				return nil, err
			}
			s := string(b)
			sf.ForwardedData = &s
		}
		if len(f.additionalData) > 0 {
			b, err := json.Marshal(f.additionalData)
			if err != nil {
				return nil, err
			}
			s := string(b)
			sf.AdditionalData = &s
		}
		serializedFrames = append(serializedFrames, sf)
	}
	b, err := json.Marshal(serializedFrames)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}

// ToEngineContext converts the database model to the flow engine context.
// resolveGraph is called to load graphs for any saved call frames; pass nil when
// you know the context was created without a frame stack (e.g. unit tests).
func (f *FlowContextDB) ToEngineContext(ctx context.Context, graph core.GraphInterface,
	resolveGraph GraphResolverFunc) (EngineContext, error) {
	var content flowContextContent
	if err := json.Unmarshal([]byte(f.Context), &content); err != nil {
		return EngineContext{}, err
	}
	// Parse user inputs
	var userInputs map[string]string
	if content.UserInputs != nil {
		if err := json.Unmarshal([]byte(*content.UserInputs), &userInputs); err != nil {
			return EngineContext{}, err
		}
	} else {
		userInputs = make(map[string]string)
	}

	// Parse runtime data
	var runtimeData map[string]string
	if content.RuntimeData != nil {
		if err := json.Unmarshal([]byte(*content.RuntimeData), &runtimeData); err != nil {
			return EngineContext{}, err
		}
	} else {
		runtimeData = make(map[string]string)
	}

	// Parse authenticated user attributes
	var userAttributes map[string]interface{}
	if content.UserAttributes != nil {
		if err := json.Unmarshal([]byte(*content.UserAttributes), &userAttributes); err != nil {
			return EngineContext{}, err
		}
	} else {
		userAttributes = make(map[string]interface{})
	}

	var token string
	if content.Token != nil {
		token = *content.Token
	}

	// Parse available attributes
	var availableAttributes *authnprovidercm.AttributesResponse
	if content.AvailableAttributes != nil && strings.TrimSpace(*content.AvailableAttributes) != "" {
		var attrs authnprovidercm.AttributesResponse
		if err := json.Unmarshal([]byte(*content.AvailableAttributes), &attrs); err != nil {
			return EngineContext{}, err
		}
		availableAttributes = &attrs
	}

	// Build authenticated user
	authenticatedUser := authncm.AuthenticatedUser{
		IsAuthenticated:     content.IsAuthenticated,
		UserID:              "",
		Attributes:          userAttributes,
		Token:               token,
		AvailableAttributes: availableAttributes,
	}
	if content.UserID != nil {
		authenticatedUser.UserID = *content.UserID
	}
	if content.OUID != nil {
		authenticatedUser.OUID = *content.OUID
	}
	if content.UserType != nil {
		authenticatedUser.UserType = *content.UserType
	}

	// Parse execution history
	var executionHistory map[string]*common.NodeExecutionRecord
	if content.ExecutionHistory != nil {
		if err := json.Unmarshal([]byte(*content.ExecutionHistory), &executionHistory); err != nil {
			return EngineContext{}, err
		}
	} else {
		executionHistory = make(map[string]*common.NodeExecutionRecord)
	}

	// Get current node from graph if available
	var currentNode core.NodeInterface
	if content.CurrentNodeID != nil {
		if node, exists := graph.GetNode(*content.CurrentNodeID); exists {
			currentNode = node
		}
	}

	// Get current action
	currentAction := ""
	if content.CurrentAction != nil {
		currentAction = *content.CurrentAction
	}

	// Get current segment ID
	currentSegmentID := ""
	if content.CurrentSegmentID != nil {
		currentSegmentID = *content.CurrentSegmentID
	}

	// Deserialize AuthUser if present
	var authUser managerpkg.AuthUser
	if content.AuthUser != nil {
		if err := json.Unmarshal([]byte(*content.AuthUser), &authUser); err != nil {
			return EngineContext{}, err
		}
	}

	// Get challenge token hash from JSON content
	challengeTokenHash := ""
	if content.ChallengeTokenHash != nil {
		challengeTokenHash = *content.ChallengeTokenHash
	}

	// Parse shared runtime data
	var sharedRuntimeData map[string]string
	if content.SharedRuntimeData != nil {
		if err := json.Unmarshal([]byte(*content.SharedRuntimeData), &sharedRuntimeData); err != nil {
			return EngineContext{}, err
		}
	}

	// Parse frame stack
	frameStack, err := deserializeFrameStack(ctx, content, resolveGraph)
	if err != nil {
		return EngineContext{}, err
	}

	return EngineContext{
		Context:            ctx,
		ExecutionID:        f.ExecutionID,
		TraceID:            "", // TraceID is transient and set from request context
		FlowType:           graph.GetType(),
		AppID:              content.AppID,
		Verbose:            content.Verbose,
		UserInputs:         userInputs,
		RuntimeData:        runtimeData,
		CurrentNode:        currentNode,
		CurrentAction:      currentAction,
		CurrentSegmentID:   currentSegmentID,
		Graph:              graph,
		AuthenticatedUser:  authenticatedUser,
		AuthUser:           authUser,
		ExecutionHistory:   executionHistory,
		ChallengeTokenHash: challengeTokenHash,
		frameStack:         frameStack,
		SharedRuntimeData:  sharedRuntimeData,
	}, nil
}

// FromEngineContext creates a database model from the flow engine context.
func FromEngineContext(ctx EngineContext) (*FlowContextDB, error) {
	// Serialize user inputs
	userInputsJSON, err := json.Marshal(ctx.UserInputs)
	if err != nil {
		return nil, err
	}
	userInputs := string(userInputsJSON)

	// Serialize runtime data
	runtimeDataJSON, err := json.Marshal(ctx.RuntimeData)
	if err != nil {
		return nil, err
	}
	runtimeData := string(runtimeDataJSON)

	// Serialize authenticated user attributes
	userAttributesJSON, err := json.Marshal(ctx.AuthenticatedUser.Attributes)
	if err != nil {
		return nil, err
	}
	userAttributes := string(userAttributesJSON)

	// Serialize execution history
	executionHistoryJSON, err := json.Marshal(ctx.ExecutionHistory)
	if err != nil {
		return nil, err
	}
	executionHistory := string(executionHistoryJSON)

	// Get current node ID
	var currentNodeID *string
	if ctx.CurrentNode != nil {
		nodeID := ctx.CurrentNode.GetID()
		currentNodeID = &nodeID
	}

	// Get current action
	var currentAction *string
	if ctx.CurrentAction != "" {
		currentAction = &ctx.CurrentAction
	}

	// Get current segment ID
	var currentSegmentID *string
	if ctx.CurrentSegmentID != "" {
		currentSegmentID = &ctx.CurrentSegmentID
	}

	// Get authenticated user ID
	var authenticatedUserID *string
	if ctx.AuthenticatedUser.UserID != "" {
		authenticatedUserID = &ctx.AuthenticatedUser.UserID
	}

	// Get organization unit ID
	var oUID *string
	if ctx.AuthenticatedUser.OUID != "" {
		oUID = &ctx.AuthenticatedUser.OUID
	}

	// Get user type
	var userType *string
	if ctx.AuthenticatedUser.UserType != "" {
		userType = &ctx.AuthenticatedUser.UserType
	}

	var token *string
	if ctx.AuthenticatedUser.Token != "" {
		token = &ctx.AuthenticatedUser.Token
	}

	// Serialize available attributes
	var availableAttributes *string
	if ctx.AuthenticatedUser.AvailableAttributes != nil {
		availableAttrsJSON, err := json.Marshal(ctx.AuthenticatedUser.AvailableAttributes)
		if err != nil {
			return nil, err
		}
		availableAttrsStr := string(availableAttrsJSON)
		availableAttributes = &availableAttrsStr
	}

	// Serialize AuthUser if present
	var authUserStr *string
	if ctx.AuthUser.IsAuthenticated() {
		authUserJSON, err := json.Marshal(&ctx.AuthUser)
		if err != nil {
			return nil, err
		}
		s := string(authUserJSON)
		authUserStr = &s
	}

	// Get graph ID
	if ctx.Graph == nil || ctx.Graph.GetID() == "" {
		return nil, fmt.Errorf("graph with a valid ID is required to persist engine context")
	}
	graphID := ctx.Graph.GetID()

	// Get challenge token hash
	var challengeTokenHash *string
	if ctx.ChallengeTokenHash != "" {
		challengeTokenHash = &ctx.ChallengeTokenHash
	}

	// Serialize frame stack
	frameStackStr, err := serializeFrameStack(ctx.frameStack)
	if err != nil {
		return nil, err
	}

	// Serialize shared runtime data
	var sharedRuntimeDataStr *string
	if len(ctx.SharedRuntimeData) > 0 {
		srdJSON, err := json.Marshal(ctx.SharedRuntimeData)
		if err != nil {
			return nil, err
		}
		s := string(srdJSON)
		sharedRuntimeDataStr = &s
	}

	content := flowContextContent{
		AppID:               ctx.AppID,
		Verbose:             ctx.Verbose,
		CurrentNodeID:       currentNodeID,
		CurrentAction:       currentAction,
		CurrentSegmentID:    currentSegmentID,
		GraphID:             graphID,
		RuntimeData:         &runtimeData,
		ExecutionHistory:    &executionHistory,
		IsAuthenticated:     ctx.AuthenticatedUser.IsAuthenticated,
		UserID:              authenticatedUserID,
		OUID:                oUID,
		UserType:            userType,
		UserInputs:          &userInputs,
		UserAttributes:      &userAttributes,
		Token:               token,
		AvailableAttributes: availableAttributes,
		AuthUser:            authUserStr,
		ChallengeTokenHash:  challengeTokenHash,
		FrameStack:          frameStackStr,
		SharedRuntimeData:   sharedRuntimeDataStr,
	}

	contextJSON, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return &FlowContextDB{
		ExecutionID: ctx.ExecutionID,
		Context:     string(contextJSON),
	}, nil
}
