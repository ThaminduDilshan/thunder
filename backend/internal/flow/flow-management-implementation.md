# Flow Management API Implementation Guide

## Overview

This document outlines the implementation approach for the Flow Management API, which provides CRUD operations for managing authentication and registration flow definitions. The system uses a dual-representation model where flows are defined using a UI-friendly step-based format and internally compiled into a graph-based execution model.

## Architecture

### Dual Representation Model

1. **Step-Based Representation (External/API)**
   - User-facing format designed for visual flow composers
   - Contains UI components, layout information (position, size)
   - Simpler, more intuitive structure for defining flows
   - Stored in database alongside graph representation

2. **Graph-Based Representation (Internal/Runtime)**
   - Execution model used by the flow engine
   - Nodes connected by edges representing execution paths
   - Automatically compiled from step definitions
   - Cached at runtime for performance

### Translation Strategy

The translator converts between step and graph representations:

**Steps → Graph (Compilation)**
- VIEW steps with multiple ACTION buttons → Generate DECISION node + multiple TASK_EXECUTION nodes
- VIEW steps with single ACTION → Single node creation
- EXECUTION steps → TASK_EXECUTION nodes
- Special steps (REGISTRATION_START, AUTHENTICATION_SUCCESS, PROVISIONING) → Specific node types

**Graph → Steps (Decompilation)**
- Reconstruct UI layout and component hierarchy from graph nodes
- Reverse DECISION node patterns back to VIEW steps with multiple actions

## Implementation Tasks

### 1. Data Models (`internal/flow/flowmgt/model.go`)

Create new structs for step-based representation:

```go
// Step-based flow definition models
type FlowDefinition struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    FlowType  string    `json:"flowType"` // AUTHENTICATION, REGISTRATION
    Steps     []Step    `json:"steps"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

type Step struct {
    ID       string       `json:"id"`
    Type     string       `json:"type"` // VIEW, EXECUTION, AUTHENTICATION_SUCCESS, REGISTRATION_START, PROVISIONING
    Size     *Size        `json:"size,omitempty"`
    Position *Position    `json:"position,omitempty"`
    Data     *StepData    `json:"data,omitempty"`
}

type Size struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

type Position struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type StepData struct {
    Components []Component `json:"components,omitempty"`
    Action     *Action     `json:"action,omitempty"`
}

type Component struct {
    ID         string       `json:"id"`
    Category   string       `json:"category"` // DISPLAY, FIELD, ACTION, BLOCK
    Type       string       `json:"type"`
    Variant    string       `json:"variant,omitempty"`
    Config     interface{}  `json:"config,omitempty"`
    Components []Component  `json:"components,omitempty"` // Nested components for BLOCK
    Action     *Action      `json:"action,omitempty"`     // For ACTION category
}

type Action struct {
    Type     string        `json:"type"` // EXECUTOR, NEXT
    Executor *ExecutorRef  `json:"executor,omitempty"`
    Next     string        `json:"next,omitempty"`
}

type ExecutorRef struct {
    Name string                 `json:"name"`
    Meta map[string]interface{} `json:"meta,omitempty"`
}
```

### 2. Flow Translator (`internal/flow/flowmgt/translator.go`)

Create translator interface and implementation:

```go
type FlowTranslatorInterface interface {
    // Convert step-based definition to graph
    StepsToGraph(definition *FlowDefinition) (*GraphDefinition, error)
    
    // Convert graph back to step-based definition
    GraphToSteps(graph *GraphDefinition) (*FlowDefinition, error)
}

type flowTranslator struct {
    factory core.FlowFactoryInterface
}

func newFlowTranslator(factory core.FlowFactoryInterface) FlowTranslatorInterface {
    return &flowTranslator{
        factory: factory,
    }
}
```

**Translation Rules:**

1. **VIEW Step with Multiple ACTIONs:**
   ```
   VIEW step (id: view_1)
   ├─ ACTION button_a (executor: ExecutorA, next: step_2)
   └─ ACTION button_b (executor: ExecutorB, next: step_3)
   
   Translates to:
   
   PROMPT_ONLY node (id: view_1) 
   → DECISION node (id: view_1_decision)
     ├─ TASK_EXECUTION node (id: view_1_action_a, executor: ExecutorA) → step_2
     └─ TASK_EXECUTION node (id: view_1_action_b, executor: ExecutorB) → step_3
   ```

2. **VIEW Step with Single ACTION:**
   ```
   VIEW step (id: view_1)
   └─ ACTION button_a (executor: ExecutorA, next: step_2)
   
   Translates to:
   
   TASK_EXECUTION node (id: view_1, executor: ExecutorA) → step_2
   ```

3. **ACTION with type=NEXT (no executor):**
   ```
   Direct edge to referenced step (no node creation)
   ```

4. **EXECUTION Step:**
   ```
   EXECUTION step (id: exec_1, executor: GoogleOIDC)
   
   Translates to:
   
   TASK_EXECUTION node (id: exec_1, executor: GoogleOIDC)
   ```

5. **Special Steps:**
   - `REGISTRATION_START` → Node with `UserTypeResolverExecutor` (required)
   - `AUTHENTICATION_SUCCESS` → Terminal node with `AuthAssertExecutor`
   - `PROVISIONING` → Node with `ProvisioningExecutor` (required for REGISTRATION flows)

### 3. Database Schema

Create new table for storing flow definitions:

```sql
-- PostgreSQL
CREATE TABLE flow_definitions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    flow_type VARCHAR(50) NOT NULL, -- AUTHENTICATION, REGISTRATION
    step_definition JSONB NOT NULL, -- Step-based representation
    graph_definition JSONB NOT NULL, -- Compiled graph representation
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT flow_type_check CHECK (flow_type IN ('AUTHENTICATION', 'REGISTRATION'))
);

CREATE INDEX idx_flow_type ON flow_definitions(flow_type);
CREATE INDEX idx_flow_name ON flow_definitions(name);

-- SQLite
CREATE TABLE flow_definitions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    flow_type TEXT NOT NULL CHECK(flow_type IN ('AUTHENTICATION', 'REGISTRATION')),
    step_definition TEXT NOT NULL,
    graph_definition TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_type ON flow_definitions(flow_type);
CREATE INDEX idx_flow_name ON flow_definitions(name);
```

### 4. Store Layer (`internal/flow/flowmgt/store.go`)

```go
type flowMgtStoreInterface interface {
    CreateFlow(ctx context.Context, flow *FlowDefinition, graphDef *GraphDefinition) error
    GetFlow(ctx context.Context, flowID string) (*FlowDefinition, *GraphDefinition, error)
    UpdateFlow(ctx context.Context, flow *FlowDefinition, graphDef *GraphDefinition) error
    DeleteFlow(ctx context.Context, flowID string) error
    ListFlows(ctx context.Context, flowType string, limit, offset int) ([]*FlowDefinition, int, error)
}

type flowMgtStore struct {
    dbClient database.DBClient
}

func newFlowMgtStore() flowMgtStoreInterface {
    return &flowMgtStore{
        dbClient: database.DBProvider.GetDBClient(),
    }
}
```

**Query Constants (`internal/flow/flowmgt/storeconstants.go`):**

```go
const (
    QueryCreateFlow = "flow.create"
    QueryGetFlow    = "flow.get"
    QueryUpdateFlow = "flow.update"
    QueryDeleteFlow = "flow.delete"
    QueryListFlows  = "flow.list"
    QueryCountFlows = "flow.count"
)

var flowMgtQueries = []model.DBQuery{
    {
        ID: QueryCreateFlow,
        Queries: map[string]string{
            "postgres": `INSERT INTO flow_definitions (id, name, flow_type, step_definition, graph_definition, created_at, updated_at) 
                        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
            "sqlite": `INSERT INTO flow_definitions (id, name, flow_type, step_definition, graph_definition, created_at, updated_at) 
                      VALUES (?, ?, ?, ?, ?, ?, ?)`,
        },
    },
    // ... other queries
}
```

### 5. Service Layer (`internal/flow/flowmgt/service.go`)

```go
type FlowManagementServiceInterface interface {
    CreateFlow(ctx context.Context, definition *FlowDefinition) (*FlowDefinition, error)
    GetFlow(ctx context.Context, flowID string) (*FlowDefinition, error)
    UpdateFlow(ctx context.Context, flowID string, definition *FlowDefinition) (*FlowDefinition, error)
    DeleteFlow(ctx context.Context, flowID string) error
    ListFlows(ctx context.Context, flowType string, limit, offset int) ([]*FlowDefinition, int, error)
}

type flowManagementService struct {
    store      flowMgtStoreInterface
    translator FlowTranslatorInterface
    validator  flowValidatorInterface
}

func newFlowManagementService(store flowMgtStoreInterface, translator FlowTranslatorInterface) FlowManagementServiceInterface {
    return &flowManagementService{
        store:      store,
        translator: translator,
        validator:  newFlowValidator(),
    }
}
```

**Service Logic:**

1. **CreateFlow:**
   - Validate flow definition (step IDs unique, required steps present, etc.)
   - Translate steps to graph using translator
   - Store both representations in database
   - Return created flow with generated ID and timestamps

2. **GetFlow:**
   - Retrieve flow from database
   - Return step-based representation only (API response)

3. **UpdateFlow:**
   - Validate updated flow definition
   - Translate to graph
   - Update database with new representations
   - Update `updated_at` timestamp

4. **DeleteFlow:**
   - Check if flow can be deleted (not default, not in use)
   - Delete from database

5. **ListFlows:**
   - Query flows with pagination and optional type filter
   - Return list with totalResults, startIndex, count

### 6. Validation (`internal/flow/flowmgt/validator.go`)

```go
type flowValidatorInterface interface {
    ValidateFlow(definition *FlowDefinition) error
}

type flowValidator struct{}

func newFlowValidator() flowValidatorInterface {
    return &flowValidator{}
}
```

**Validation Rules:**

- All step IDs must be unique within the flow
- All `next` references must point to valid step IDs
- AUTHENTICATION flows must contain an AUTHENTICATION_SUCCESS step
- REGISTRATION flows must:
  - Start with REGISTRATION_START step (first step)
  - Contain a PROVISIONING step
  - End with AUTHENTICATION_SUCCESS step
- Special steps must have required executors:
  - REGISTRATION_START → UserTypeResolverExecutor
  - PROVISIONING → ProvisioningExecutor
  - AUTHENTICATION_SUCCESS → AuthAssertExecutor (for REGISTRATION flows)
- VIEW steps must have at least one ACTION component
- i18n keys should follow dot notation format (validation can be lenient)

### 7. HTTP Handler Layer (`internal/flow/flowmgt/handler.go`)

```go
type flowMgtHandler struct {
    service FlowManagementServiceInterface
}

func newFlowMgtHandler(service FlowManagementServiceInterface) *flowMgtHandler {
    return &flowMgtHandler{
        service: service,
    }
}

func (h *flowMgtHandler) handleCreateFlow(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    // Call service.CreateFlow
    // Return 201 with created flow
}

func (h *flowMgtHandler) handleGetFlow(w http.ResponseWriter, r *http.Request) {
    // Extract flowId from path
    // Call service.GetFlow
    // Return 200 with flow
}

func (h *flowMgtHandler) handleUpdateFlow(w http.ResponseWriter, r *http.Request) {
    // Extract flowId from path
    // Parse request body
    // Call service.UpdateFlow
    // Return 200 with updated flow
}

func (h *flowMgtHandler) handleDeleteFlow(w http.ResponseWriter, r *http.Request) {
    // Extract flowId from path
    // Call service.DeleteFlow
    // Return 204 on success
}

func (h *flowMgtHandler) handleListFlows(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters (flowType, limit, offset)
    // Call service.ListFlows
    // Return 200 with paginated list
}
```

### 8. Route Registration (`internal/flow/flowmgt/init.go`)

```go
func Initialize(mux *http.ServeMux) {
    // Create store
    store := newFlowMgtStore()
    
    // Create translator
    factory := core.GetFlowFactory()
    translator := newFlowTranslator(factory)
    
    // Create service
    service := newFlowManagementService(store, translator)
    
    // Create handler
    handler := newFlowMgtHandler(service)
    
    // Register routes
    registerRoutes(mux, handler)
}

func registerRoutes(mux *http.ServeMux, handler *flowMgtHandler) {
    corsPolicy := middleware.GetDefaultCORSPolicy()
    
    mux.Handle("GET /flows", middleware.WithCORS(
        http.HandlerFunc(handler.handleListFlows), corsPolicy))
    
    mux.Handle("POST /flows", middleware.WithCORS(
        http.HandlerFunc(handler.handleCreateFlow), corsPolicy))
    
    mux.Handle("GET /flows/{flowId}", middleware.WithCORS(
        http.HandlerFunc(handler.handleGetFlow), corsPolicy))
    
    mux.Handle("PUT /flows/{flowId}", middleware.WithCORS(
        http.HandlerFunc(handler.handleUpdateFlow), corsPolicy))
    
    mux.Handle("DELETE /flows/{flowId}", middleware.WithCORS(
        http.handlerFunc(handler.handleDeleteFlow), corsPolicy))
}
```

### 9. Error Constants (`internal/flow/flowmgt/errorconstants.go`)

```go
const (
    // API Error Codes
    ErrorCodeInvalidRequest      = "FMS-1001"
    ErrorCodeInvalidFlowID       = "FMS-1002"
    ErrorCodeInvalidStepID       = "FMS-1003"
    ErrorCodeMissingRequiredStep = "FMS-1004"
    ErrorCodeInvalidExecutor     = "FMS-1005"
    ErrorCodeFlowNotFound        = "FMS-1006"
    ErrorCodeFlowCannotBeDeleted = "FMS-1007"
)

const (
    // Error Messages
    ErrorMessageInvalidRequest      = "Invalid request format"
    ErrorMessageInvalidFlowID       = "Invalid flow ID"
    ErrorMessageFlowNotFound        = "Flow not found"
    ErrorMessageFlowCannotBeDeleted = "Flow cannot be deleted"
)
```

### 10. Integration with Existing Flow Engine

**Runtime Graph Compilation:**

The existing flow execution engine will need to:

1. Load flow definition from database (graph representation)
2. Cache compiled graph in memory for performance
3. Invalidate cache when flow is updated

**Service Manager Update (`cmd/server/servicemanager.go`):**

```go
func registerServices(mux *http.ServeMux) {
    // ... existing services
    
    // Initialize Flow Management
    flowmgt.Initialize(mux)
    
    // ... other services
}
```

## Testing Strategy

### Unit Tests

1. **Translator Tests:**
   - Test each translation rule (VIEW → DECISION, EXECUTION, special steps)
   - Test reverse translation (Graph → Steps)
   - Test edge cases (empty flows, invalid references)

2. **Validator Tests:**
   - Test all validation rules
   - Test flow type-specific requirements
   - Test error scenarios

3. **Service Tests:**
   - Mock store layer
   - Test CRUD operations
   - Test error handling

4. **Store Tests:**
   - Use `go-sqlmock` to test database operations
   - Test query execution for both PostgreSQL and SQLite
   - Test error scenarios

### Integration Tests

1. **API Tests (`tests/integration/flowmgt/`):**
   - Test complete CRUD flow via HTTP
   - Test pagination and filtering
   - Test validation errors
   - Test flow execution after creation/update

2. **End-to-End Tests:**
   - Create flow via API
   - Execute flow via existing flow engine
   - Verify execution results

## Migration Plan

### Phase 1: Core Implementation
1. Implement data models
2. Implement translator (Steps → Graph)
3. Implement basic validator
4. Add database table and migrations
5. Implement store layer

### Phase 2: Service & API
1. Implement service layer
2. Implement HTTP handlers
3. Register routes
4. Add error handling

### Phase 3: Testing & Refinement
1. Write unit tests (target: 80%+ coverage)
2. Write integration tests
3. Test with existing flow engine
4. Performance optimization (caching, indexing)

### Phase 4: Documentation & Examples
1. Update API documentation
2. Add example flows
3. Create user guide for flow creation
4. Add migration guide for existing flows

## Future Enhancements

1. **Flow Versioning:**
   - Track flow definition versions
   - Allow rollback to previous versions
   - Show version history

2. **Flow Templates:**
   - Pre-defined flow templates
   - Template marketplace

3. **Visual Flow Composer UI:**
   - Drag-and-drop flow builder
   - Real-time validation
   - Preview mode

4. **Flow Analytics:**
   - Track flow execution metrics
   - Identify bottlenecks
   - A/B testing support

5. **Advanced Validation:**
   - Circular dependency detection
   - Dead code detection
   - Accessibility validation for UI components

## References

- OpenAPI Specification: `/api/WIP/flow-management.yaml`
- Existing Flow Engine: `/backend/internal/flow/`
- Database Scripts: `/backend/dbscripts/`
- Integration Tests: `/tests/integration/`
