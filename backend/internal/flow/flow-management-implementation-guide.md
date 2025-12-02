# Flow Management API - Implementation Guide

This document provides implementation details for the Flow Management API based on the [Revised Approach](./flow-management-revised-approach.md).

## Backend Implementation

### 1. Update Node Type Constants

**File:** `backend/internal/flow/common/constants.go`

```go
const (
    // NodeTypeUIInteraction represents a node that handles user interface interactions
    NodeTypeUIInteraction NodeType = "UI_INTERACTION"
    
    // Existing node types...
    NodeTypeAuthSuccess NodeType = "AUTHENTICATION_SUCCESS"
    NodeTypeRegistrationStart NodeType = "REGISTRATION_START"
    NodeTypeTaskExecution NodeType = "TASK_EXECUTION"
    NodeTypePromptOnly NodeType = "PROMPT_ONLY"
    NodeTypeDecision NodeType = "DECISION"
    NodeTypeProvisioning NodeType = "PROVISIONING"
)
```

### 2. Update Graph Models

**File:** `backend/internal/flow/flowmgt/model.go`

```go
type graphDefinition struct {
    ID    string           `json:"id"`
    Type  string           `json:"type"`
    Nodes []nodeDefinition `json:"nodes"`
}

type nodeDefinition struct {
    ID           string                 `json:"id"`
    Type         string                 `json:"type"`
    Properties   map[string]interface{} `json:"properties,omitempty"`  // For executor-specific props (idpId, etc)
    InputData    []inputDefinition      `json:"inputData,omitempty"`
    Executor     *executorDefinition    `json:"executor,omitempty"`
    UIDefinition *uiDefinition          `json:"uiDefinition"`           // Required for all nodes (contains layout)
    Next         []string               `json:"next,omitempty"`
}

type uiDefinition struct {
    Layout     *uiLayout     `json:"layout"`            // Required for all nodes (composer positioning)
    Components []uiComponent `json:"components,omitempty"` // Only for UI_INTERACTION nodes
}

type uiLayout struct {
    Size     *uiSize     `json:"size"`
    Position *uiPosition `json:"position"`
}

type uiSize struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

type uiPosition struct {
    X float64 `json:"x"`
    Y float64 `json:"y"`
}

type uiComponent struct {
    ID         string                 `json:"id"`
    Category   string                 `json:"category"` // DISPLAY, FIELD, ACTION, BLOCK
    Type       string                 `json:"type"`
    Variant    string                 `json:"variant,omitempty"`
    Config     map[string]interface{} `json:"config,omitempty"`
    Components []uiComponent          `json:"components,omitempty"` // Nested components for BLOCK category
    Action     *uiAction              `json:"action,omitempty"`     // Action for ACTION category components
}

type uiAction struct {
    Type string `json:"type"` // "NEXT" - for navigation
    Next string `json:"next"` // Target node ID
}

type executorDefinition struct {
    Name string `json:"name"`
}
```

### 3. Flow Management Service

**File:** `backend/internal/flow/flowmgt/service.go`

**No translation needed!** Service is much simpler:

```go
type FlowManagementServiceInterface interface {
    CreateFlow(ctx context.Context, definition *graphDefinition) (*FlowResponse, error)
    GetFlow(ctx context.Context, flowID string) (*FlowResponse, error)
    UpdateFlow(ctx context.Context, flowID string, definition *graphDefinition) (*FlowResponse, error)
    DeleteFlow(ctx context.Context, flowID string) error
    ListFlows(ctx context.Context, flowType string, limit, offset int) ([]*FlowListItem, int, error)
}

type FlowResponse struct {
    ID          string           `json:"id"`
    Name        string           `json:"name"`
    Type        string           `json:"type"`
    Nodes       []nodeDefinition `json:"nodes"`
    CreatedAt   string           `json:"createdAt"`
    UpdatedAt   string           `json:"updatedAt"`
}

func (s *flowManagementService) CreateFlow(ctx context.Context, definition *graphDefinition) (*FlowResponse, error) {
    // 1. Validate graph structure
    if err := s.validator.ValidateGraph(definition); err != nil {
        return nil, err
    }
    
    // 2. Generate ID and timestamps
    flowID := generateID()
    now := time.Now()
    
    // 3. Store graph as-is (no translation!)
    if err := s.store.CreateFlow(ctx, flowID, definition, now); err != nil {
        return nil, err
    }
    
    // 4. Return response
    return &FlowResponse{
        ID:        flowID,
        Name:      definition.ID,
        Type:      definition.Type,
        Nodes:     definition.Nodes,
        CreatedAt: now.Format(time.RFC3339),
        UpdatedAt: now.Format(time.RFC3339),
    }, nil
}
```

### 4. Validation Rules

**File:** `backend/internal/flow/flowmgt/validator.go`

```go
type graphValidatorInterface interface {
    ValidateGraph(definition *graphDefinition) error
}

type graphValidator struct{}

func newGraphValidator() graphValidatorInterface {
    return &graphValidator{}
}

func (v *graphValidator) ValidateGraph(definition *graphDefinition) error {
    // 1. Validate node IDs are unique
    if err := v.validateUniqueNodeIDs(definition); err != nil {
        return err
    }
    
    // 2. Validate all "next" references point to valid node IDs
    if err := v.validateNextReferences(definition); err != nil {
        return err
    }
    
    // 3. Validate flow type-specific requirements
    if err := v.validateFlowTypeRequirements(definition); err != nil {
        return err
    }
    
    // 4. Validate each node
    for _, node := range definition.Nodes {
        if err := v.validateNode(&node, definition); err != nil {
            return err
        }
    }
    
    return nil
}

func (v *graphValidator) validateNode(node *nodeDefinition, graph *graphDefinition) error {
    // Validate layout
    if err := validateLayouts(node); err != nil {
        return err
    }
    
    // Node-specific validation
    switch node.Type {
    case "UI_INTERACTION":
        return validateUIInteractionNode(node, graph)
    case "TASK_EXECUTION":
        return validateTaskExecutionNode(node)
    case "DECISION":
        return validateDecisionNode(node)
    // ... other node types
    }
    
    return nil
}
```

#### Multiple Actions Require DECISION Node

```go
func validateUIInteractionNode(node *nodeDefinition, graph *graphDefinition) error {
    if node.UIDefinition == nil {
        return errors.New("UI_INTERACTION node must have uiDefinition")
    }
    
    // Extract all ACTION components (recursively from nested components)
    actions := extractActionComponents(node.UIDefinition.Components)
    actionCount := len(actions)
    
    if actionCount == 0 {
        return errors.New("UI_INTERACTION must have at least one ACTION component")
    }
    
    if actionCount == 1 {
        // Single action: can directly link to target
        if len(node.Next) != 1 {
            return errors.New("UI_INTERACTION with single action must have exactly one next node")
        }
    } else {
        // Multiple actions: must have DECISION node next
        if len(node.Next) != 1 {
            return errors.New("UI_INTERACTION with multiple actions must have exactly one next node (DECISION)")
        }
        
        nextNode := findNodeByID(graph, node.Next[0])
        if nextNode.Type != "DECISION" {
            return errors.New("UI_INTERACTION with multiple actions must link to DECISION node")
        }
        
        // Validate DECISION includes all action targets
        actionTargets := extractActionNextTargets(actions)
        if !allTargetsInDecision(actionTargets, nextNode.Next) {
            return errors.New("DECISION node must include all action target nodes")
        }
        
        // Validate all actions have type NEXT
        for _, action := range actions {
            if action.Action == nil || action.Action.Type != "NEXT" {
                return errors.New("All ACTION components must have action.type = NEXT")
            }
        }
    }
    
    return nil
}

// Helper function to recursively extract ACTION components
func extractActionComponents(components []uiComponent) []uiComponent {
    var actions []uiComponent
    for _, comp := range components {
        if comp.Category == "ACTION" {
            actions = append(actions, comp)
        }
        // Recursively check nested components (for BLOCK types like FORM)
        if len(comp.Components) > 0 {
            actions = append(actions, extractActionComponents(comp.Components)...)
        }
    }
    return actions
}
```

#### Layout Validation

```go
func validateLayouts(node *nodeDefinition) error {
    // ALL nodes must have uiDefinition.layout for flow composer positioning
    if node.UIDefinition == nil || node.UIDefinition.Layout == nil {
        return errors.New("All nodes must have uiDefinition.layout for flow composer")
    }
    
    layout := node.UIDefinition.Layout
    
    // Validate size structure
    if layout.Size == nil {
        return errors.New("uiDefinition.layout.size is required")
    }
    if layout.Size.Width <= 0 || layout.Size.Height <= 0 {
        return errors.New("uiDefinition.layout.size must have positive width and height")
    }
    
    // Validate position structure
    if layout.Position == nil {
        return errors.New("uiDefinition.layout.position is required")
    }
    // Note: x and y can be negative (nodes can be placed anywhere on canvas)
    
    // UI_INTERACTION nodes must have components
    if node.Type == "UI_INTERACTION" {
        if node.UIDefinition.Components == nil || len(node.UIDefinition.Components) == 0 {
            return errors.New("UI_INTERACTION node must have components")
        }
    }
    
    return nil
}
```

### 5. Database Schema

**PostgreSQL:**
```sql
CREATE TABLE flow_definitions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    flow_type VARCHAR(50) NOT NULL,
    graph_definition JSONB NOT NULL,  -- Store complete graph as-is
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT flow_type_check CHECK (flow_type IN ('AUTHENTICATION', 'REGISTRATION'))
);

CREATE INDEX idx_flow_type ON flow_definitions(flow_type);
CREATE INDEX idx_flow_name ON flow_definitions(name);
```

**SQLite:**
```sql
CREATE TABLE flow_definitions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    flow_type TEXT NOT NULL CHECK(flow_type IN ('AUTHENTICATION', 'REGISTRATION')),
    graph_definition TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_flow_type ON flow_definitions(flow_type);
CREATE INDEX idx_flow_name ON flow_definitions(name);
```

**Note:** We only store ONE representation (graph), not two!

### 6. Store Layer

**File:** `backend/internal/flow/flowmgt/store.go`

```go
type flowMgtStoreInterface interface {
    CreateFlow(ctx context.Context, flowID string, definition *graphDefinition, timestamp time.Time) error
    GetFlow(ctx context.Context, flowID string) (*graphDefinition, time.Time, time.Time, error)
    UpdateFlow(ctx context.Context, flowID string, definition *graphDefinition, timestamp time.Time) error
    DeleteFlow(ctx context.Context, flowID string) error
    ListFlows(ctx context.Context, flowType string, limit, offset int) ([]*FlowListItem, int, error)
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
            "postgres": `INSERT INTO flow_definitions (id, name, flow_type, graph_definition, created_at, updated_at) 
                        VALUES ($1, $2, $3, $4, $5, $6)`,
            "sqlite": `INSERT INTO flow_definitions (id, name, flow_type, graph_definition, created_at, updated_at) 
                      VALUES (?, ?, ?, ?, ?, ?)`,
        },
    },
    {
        ID: QueryGetFlow,
        Queries: map[string]string{
            "postgres": `SELECT graph_definition, created_at, updated_at FROM flow_definitions WHERE id = $1`,
            "sqlite": `SELECT graph_definition, created_at, updated_at FROM flow_definitions WHERE id = ?`,
        },
    },
    // ... other queries
}
```

### 7. UI Interaction Node Executor

**File:** `backend/internal/flow/core/ui_interaction_node.go` (NEW)

```go
package core

import (
    "github.com/ThaminduDilshan/thunder/backend/internal/flow/common"
)

// UIInteractionNode handles UI interaction nodes in the flow
type uiInteractionNode struct {
    baseNode
    uiDefinition *uiDefinition
}

func newUIInteractionNode(id string, uiDef *uiDefinition, next []string) NodeInterface {
    return &uiInteractionNode{
        baseNode: baseNode{
            id:   id,
            next: next,
        },
        uiDefinition: uiDef,
    }
}

func (n *uiInteractionNode) Execute(ctx *NodeContext) (*common.NodeResponse, error) {
    // Return UI definition to frontend for rendering
    return &common.NodeResponse{
        Status:       common.NodeStatusIncomplete,
        Type:         common.NodeResponseTypeView,
        RequiredData: extractRequiredFields(n.uiDefinition),
        AdditionalData: map[string]interface{}{
            "uiDefinition": n.uiDefinition,
        },
        NextNodeID: determineNextNode(ctx, n.next),
    }, nil
}

func extractRequiredFields(uiDef *uiDefinition) []string {
    // Recursively extract identifiers from FIELD components
    var fields []string
    for _, comp := range uiDef.Components {
        if comp.Category == "FIELD" {
            if identifier, ok := comp.Config["identifier"].(string); ok {
                fields = append(fields, identifier)
            }
        }
        // Check nested components
        if len(comp.Components) > 0 {
            fields = append(fields, extractFieldsFromComponents(comp.Components)...)
        }
    }
    return fields
}
```

**Register in Factory (`internal/flow/core/factory.go`):**

```go
func (f *flowFactory) CreateNode(nodeDef *nodeDefinition) (NodeInterface, error) {
    switch nodeDef.Type {
    case "UI_INTERACTION":
        return newUIInteractionNode(nodeDef.ID, nodeDef.UIDefinition, nodeDef.Next), nil
    case "TASK_EXECUTION":
        // ... existing logic
    // ... other node types
    }
}
```

### 8. HTTP Handler Layer

**File:** `backend/internal/flow/flowmgt/handler.go`

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
    var definition graphDefinition
    if err := json.NewDecoder(r.Body).Decode(&definition); err != nil {
        apierror.WriteError(w, apierror.NewBadRequestError("Invalid request body"))
        return
    }
    
    response, err := h.service.CreateFlow(r.Context(), &definition)
    if err != nil {
        apierror.WriteServiceError(w, err)
        return
    }
    
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

func (h *flowMgtHandler) handleGetFlow(w http.ResponseWriter, r *http.Request) {
    flowID := r.PathValue("flowId")
    
    response, err := h.service.GetFlow(r.Context(), flowID)
    if err != nil {
        apierror.WriteServiceError(w, err)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(response)
}

// ... other handlers (UpdateFlow, DeleteFlow, ListFlows)
```

### 9. Route Registration

**File:** `backend/internal/flow/flowmgt/init.go`

```go
func Initialize(mux *http.ServeMux) {
    // Create store
    store := newFlowMgtStore()
    
    // Create validator
    validator := newGraphValidator()
    
    // Create service
    service := newFlowManagementService(store, validator)
    
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
        http.HandlerFunc(handler.handleDeleteFlow), corsPolicy))
}
```

### 10. Error Constants

**File:** `backend/internal/flow/flowmgt/errorconstants.go`

```go
const (
    // API Error Codes
    ErrorCodeInvalidRequest      = "FMS-1001"
    ErrorCodeInvalidFlowID       = "FMS-1002"
    ErrorCodeInvalidNodeID       = "FMS-1003"
    ErrorCodeMissingRequiredNode = "FMS-1004"
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

### 11. Service Manager Integration

**File:** `cmd/server/servicemanager.go`

```go
func registerServices(mux *http.ServeMux) {
    // ... existing services
    
    // Initialize Flow Management
    flowmgt.Initialize(mux)
    
    // ... other services
}
```

## Implementation Roadmap

### Phase 1: Backend Core
1. Update constants with UI_INTERACTION node type
2. Implement graph models (graphDefinition, nodeDefinition, uiDefinition, etc.)
3. Implement validator with all validation rules
4. Create database schema and migrations
5. Implement store layer with CRUD operations
6. Implement service layer (no translator needed!)

### Phase 2: Flow Execution
1. Implement UI Interaction node executor
2. Register node type in factory
3. Update NodeResponse to support UI definitions
4. Test flow execution with UI nodes

### Phase 3: API Layer
1. Implement HTTP handlers (CREATE, GET, UPDATE, DELETE, LIST)
2. Register routes with middleware
3. Define error constants
4. Wire up in service manager

### Phase 4: Testing
1. Write unit tests (validator, service, store, executor)
2. Write integration tests (API endpoints, flow execution)
3. Achieve 80%+ test coverage

### Phase 5: Documentation
1. Update OpenAPI specification
2. Create example flow definitions
3. Document validation rules
4. Create developer guide

## Testing Strategy

### Unit Tests

**Validator Tests:**
- Test unique node ID validation
- Test next reference validation
- Test UI_INTERACTION with single action
- Test UI_INTERACTION with multiple actions requires DECISION
- Test DECISION includes all action targets
- Test layout validation (size, position)
- Test flow type-specific requirements

**Service Tests:**
- Mock store layer
- Test CreateFlow validation and storage
- Test GetFlow retrieval
- Test UpdateFlow validation and update
- Test DeleteFlow
- Test ListFlows pagination
- Test error handling

**Store Tests:**
- Use `go-sqlmock` for database operations
- Test all CRUD queries for PostgreSQL
- Test all CRUD queries for SQLite
- Test error scenarios

**Executor Tests:**
- Test UI_INTERACTION node execution
- Test UI definition returned in response
- Test required fields extraction

### Integration Tests

**File:** `tests/integration/flowmgt/`

```go
func TestFlowManagementAPI(t *testing.T) {
    // Test CREATE flow
    // Test GET flow
    // Test UPDATE flow
    // Test LIST flows
    // Test DELETE flow
}

func TestFlowExecutionWithUINodes(t *testing.T) {
    // Create flow with UI_INTERACTION nodes
    // Execute flow
    // Verify UI definition returned
    // Submit user input
    // Verify flow continues
}
```

## References

- Design Document: [flow-management-revised-approach.md](./flow-management-revised-approach.md)
- Initial Approach: [flow-management-implementation.md](./flow-management-implementation.md)
- OpenAPI Specification: `/api/WIP/flow-management.yaml`
- Existing Graph Examples: `/backend/cmd/server/repository/resources/graphs/`
