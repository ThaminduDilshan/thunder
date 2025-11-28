# Flow Management API - Revised Approach

## Problem Statement

The initial approach had a **dual-representation model** where:
- Frontend sends step-based UI definitions
- Backend translates to graph representation
- Creates complexity in translation logic and maintaining two representations

## New Approach: Single Graph Representation

### Key Changes

1. **Frontend and Backend share the same graph representation** (JSON format)
2. **Introduce a new UI node type** to encapsulate UI metadata
3. **No translation layer needed** - what frontend sends is what backend stores
4. **Frontend is responsible for graph construction** including decision nodes

### Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    Flow Composer UI (Frontend)               │
│  - Visual drag-and-drop interface                            │
│  - Constructs complete graph with UI nodes                   │
│  - Inserts decision nodes when multiple options exist        │
└────────────────────┬─────────────────────────────────────────┘
                     │ Graph JSON
                     ▼
┌──────────────────────────────────────────────────────────────┐
│              Flow Management API (Backend)                   │
│  - Stores graph definition as-is (no translation)            │
│  - Validates graph structure                                 │
│  - Serves graph for runtime execution                        │
└────────────────────┬─────────────────────────────────────────┘
                     │ Same Graph JSON
                     ▼
┌──────────────────────────────────────────────────────────────┐
│              Flow Execution Engine (Backend)                 │
│  - Loads graph from database                                 │
│  - Executes nodes including UI nodes                         │
│  - Returns UI metadata when UI node is encountered           │
└──────────────────────────────────────────────────────────────┘
```

## New Node Type: UI_INTERACTION

**Decided**: Use `UI_INTERACTION` as the node type name.
- Clear purpose: handles user interaction
- Consistent with existing naming (TASK_EXECUTION, AUTHENTICATION_SUCCESS)
- Can be renamed later if needed

### Node Layout Requirements

**All nodes** in the graph must have `properties.layout` for positioning in the flow composer:
- Position (x, y) where the node is placed in the canvas
- Size (width, height) of the node box in the composer
- Users can drag and resize nodes in the flow composer UI

**UI_INTERACTION nodes** additionally have `uiDefinition.layout` for end-user UI:
- Size (width, height) of the actual UI dialog/form rendered to end users
- No position needed (centered or positioned by frontend rendering logic)

### UI_INTERACTION Node Structure

```json
{
    "id": "login_view",
    "type": "UI_INTERACTION",
    "properties": {
        "layout": {
            "width": 200,
            "height": 100,
            "x": 100,
            "y": 200
        }
    },
    "uiDefinition": {
        "layout": {
            "width": 400,
            "height": 500
        },
        "components": [
            {
                "id": "title_1",
                "category": "DISPLAY",
                "type": "TYPOGRAPHY",
                "variant": "H3",
                "config": {
                    "text": "auth.login.title"
                }
            },
            {
                "id": "form_credentials",
                "category": "BLOCK",
                "type": "FORM",
                "config": {},
                "components": [
                    {
                        "id": "input_username",
                        "category": "FIELD",
                        "type": "INPUT",
                        "variant": "TEXT",
                        "config": {
                            "type": "text",
                            "label": "auth.username.label",
                            "placeholder": "auth.username.placeholder",
                            "required": true,
                            "identifier": "username"
                        }
                    },
                    {
                        "id": "input_password",
                        "category": "FIELD",
                        "type": "INPUT",
                        "variant": "PASSWORD",
                        "config": {
                            "type": "password",
                            "label": "auth.password.label",
                            "placeholder": "auth.password.placeholder",
                            "required": true,
                            "identifier": "password"
                        }
                    },
                    {
                        "id": "btn_continue",
                        "category": "ACTION",
                        "type": "BUTTON",
                        "variant": "PRIMARY",
                        "config": {
                            "type": "submit",
                            "text": "auth.continue.button"
                        },
                        "action": {
                            "targetNodeId": "basic_auth"
                        }
                    }
                ]
            },
            {
                "id": "btn_google",
                "category": "ACTION",
                "type": "BUTTON",
                "variant": "SOCIAL",
                "config": {
                    "type": "button",
                    "text": "auth.google.button",
                    "image": "assets/images/icons/google.svg"
                },
                "action": {
                    "targetNodeId": "google_auth"
                }
            }
        ]
    },
    "next": ["choose_method"]
}
```

## Example: Authentication Flow with Basic + Google

### Flow Structure

```
UI_INTERACTION (login_view)
    ↓
DECISION (choose_method) 
    ├─→ TASK_EXECUTION (basic_auth) → AUTHENTICATION_SUCCESS (authenticated)
    └─→ TASK_EXECUTION (google_auth) → AUTHENTICATION_SUCCESS (authenticated)
```

### Complete Graph JSON

```json
{
    "id": "auth_flow_basic_google",
    "type": "AUTHENTICATION",
    "nodes": [
        {
            "id": "login_view",
            "type": "UI_INTERACTION",
            "properties": {
                "layout": {
                    "width": 200,
                    "height": 100,
                    "x": 100,
                    "y": 200
                }
            },
            "uiDefinition": {
                "layout": {
                    "width": 400,
                    "height": 500
                },
                "components": [
                    {
                        "id": "title_1",
                        "category": "DISPLAY",
                        "type": "TYPOGRAPHY",
                        "variant": "H3",
                        "config": {
                            "text": "auth.login.title"
                        }
                    },
                    {
                        "id": "form_credentials",
                        "category": "BLOCK",
                        "type": "FORM",
                        "config": {},
                        "components": [
                            {
                                "id": "input_username",
                                "category": "FIELD",
                                "type": "INPUT",
                                "variant": "TEXT",
                                "config": {
                                    "type": "text",
                                    "label": "auth.username.label",
                                    "placeholder": "auth.username.placeholder",
                                    "required": true,
                                    "identifier": "username"
                                }
                            },
                            {
                                "id": "input_password",
                                "category": "FIELD",
                                "type": "INPUT",
                                "variant": "PASSWORD",
                                "config": {
                                    "type": "password",
                                    "label": "auth.password.label",
                                    "placeholder": "auth.password.placeholder",
                                    "required": true,
                                    "identifier": "password"
                                }
                            },
                            {
                                "id": "btn_continue",
                                "category": "ACTION",
                                "type": "BUTTON",
                                "variant": "PRIMARY",
                                "config": {
                                    "type": "submit",
                                    "text": "auth.continue.button"
                                },
                                "action": {
                                    "targetNodeId": "basic_auth"
                                }
                            }
                        ]
                    },
                    {
                        "id": "btn_google",
                        "category": "ACTION",
                        "type": "BUTTON",
                        "variant": "SOCIAL",
                        "config": {
                            "type": "button",
                            "text": "auth.google.button",
                            "image": "assets/images/icons/google.svg"
                        },
                        "action": {
                            "targetNodeId": "google_auth"
                        }
                    }
                ]
            },
            "next": ["choose_method"]
        },
        {
            "id": "choose_method",
            "type": "DECISION",
            "properties": {
                "layout": { "width": 150, "height": 80, "x": 100, "y": 400 }
            },
            "next": ["basic_auth", "google_auth"]
        },
        {
            "id": "basic_auth",
            "type": "TASK_EXECUTION",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 50, "y": 550 }
            },
            "executor": {
                "name": "BasicAuthExecutor"
            },
            "next": ["authenticated"]
        },
        {
            "id": "google_auth",
            "type": "TASK_EXECUTION",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 300, "y": 550 },
                "idpId": "<google-idp-id>"
            },
            "executor": {
                "name": "GoogleOIDCAuthExecutor"
            },
            "next": ["authenticated"]
        },
        {
            "id": "authenticated",
            "type": "AUTHENTICATION_SUCCESS",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 175, "y": 700 }
            }
        }
    ]
}
```

## How Frontend Constructs the Graph

### Scenario 1: Single Authentication Option (Basic Only)

**Visual in Flow Composer:**
```
┌─────────────────────────┐
│   Login Form            │
│   [Username]            │
│   [Password]            │
│   [Continue Button]     │
└─────────────────────────┘
           ↓
┌─────────────────────────┐
│   Basic Auth Executor   │
└─────────────────────────┘
           ↓
┌─────────────────────────┐
│   Authenticated         │
└─────────────────────────┘
```

**Generated Graph:**
```json
{
    "nodes": [
        {
            "id": "login_view",
            "type": "UI_INTERACTION",
            "uiDefinition": { /* form with username, password, continue button */ },
            "next": ["basic_auth"]  // Direct link, no decision node
        },
        {
            "id": "basic_auth",
            "type": "TASK_EXECUTION",
            "executor": { "name": "BasicAuthExecutor" },
            "next": ["authenticated"]
        },
        {
            "id": "authenticated",
            "type": "AUTHENTICATION_SUCCESS"
        }
    ]
}
```

### Scenario 2: Multiple Authentication Options (Basic + Google)

**Visual in Flow Composer:**
```
┌─────────────────────────┐
│   Login Form            │
│   [Username]            │
│   [Password]            │
│   [Continue Button] ────┼──┐
│   [Google Button]   ────┼──┤
└─────────────────────────┘  │
                             │
           ┌─────────────────┴──────────────┐
           ↓                                ↓
┌─────────────────────────┐    ┌─────────────────────────┐
│   Basic Auth Executor   │    │   Google Auth Executor  │
└─────────────────────────┘    └─────────────────────────┘
           ↓                                ↓
           └────────────────┬───────────────┘
                            ↓
                 ┌─────────────────────────┐
                 │   Authenticated         │
                 └─────────────────────────┘
```

**Generated Graph:**
```json
{
    "nodes": [
        {
            "id": "login_view",
            "type": "UI_INTERACTION",
            "uiDefinition": { 
                /* form with username, password */
                "actions": [
                    { "id": "btn_continue", "targetNodeId": "basic_auth" },
                    { "id": "btn_google", "targetNodeId": "google_auth" }
                ]
            },
            "next": ["choose_method"]  // Links to decision node
        },
        {
            "id": "choose_method",
            "type": "DECISION",
            "next": ["basic_auth", "google_auth"]
        },
        // ... basic_auth and google_auth nodes ...
    ]
}
```

**Frontend Logic:**
1. User adds UI node with form components
2. User adds action buttons (continue, google)
3. User connects buttons to target executor nodes
4. **Frontend automatically inserts DECISION node** when it detects multiple action targets
5. Frontend generates the complete graph JSON

## Example: Registration Flow

### Flow Structure

```
REGISTRATION_START (reg_start)
    ↓
UI_INTERACTION (choose_method_view)
    ↓
DECISION (choose_method)
    ├─→ TASK_EXECUTION (basic_auth) → UI_INTERACTION (profile_view) → PROVISIONING → AUTH_SUCCESS
    └─→ TASK_EXECUTION (google_auth) ─────────────────────────────────→ PROVISIONING → AUTH_SUCCESS
```

### Complete Graph JSON

```json
{
    "id": "registration_flow_basic_google",
    "type": "REGISTRATION",
    "nodes": [
        {
            "id": "reg_start",
            "type": "REGISTRATION_START",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 100, "y": 100 }
            },
            "executor": {
                "name": "UserTypeResolverExecutor"
            },
            "next": ["choose_method_view"]
        },
        {
            "id": "choose_method_view",
            "type": "UI_INTERACTION",
            "properties": {
                "layout": { "width": 200, "height": 100, "x": 100, "y": 250 }
            },
            "uiDefinition": {
                "layout": { "width": 400, "height": 500 },
                "components": [ /* username, password fields */ ],
                "actions": [
                    { "id": "btn_continue", "targetNodeId": "basic_auth" },
                    { "id": "btn_google", "targetNodeId": "google_auth" }
                ]
            },
            "next": ["choose_method"]
        },
        {
            "id": "choose_method",
            "type": "DECISION",
            "properties": {
                "layout": { "width": 150, "height": 80, "x": 100, "y": 400 }
            },
            "next": ["basic_auth", "google_auth"]
        },
        {
            "id": "basic_auth",
            "type": "TASK_EXECUTION",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 50, "y": 550 }
            },
            "executor": { "name": "BasicAuthExecutor" },
            "next": ["profile_view"]
        },
        {
            "id": "profile_view",
            "type": "UI_INTERACTION",
            "properties": {
                "layout": { "width": 200, "height": 100, "x": 100, "y": 700 }
            },
            "uiDefinition": {
                "layout": { "width": 400, "height": 500 },
                "components": [ /* email, firstName, lastName fields */ ],
                "actions": [
                    { "id": "btn_submit", "targetNodeId": "provisioning" }
                ]
            },
            "next": ["provisioning"]
        },
        {
            "id": "google_auth",
            "type": "TASK_EXECUTION",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 300, "y": 550 },
                "idpId": "<google-idp-id>"
            },
            "executor": { "name": "GoogleOIDCAuthExecutor" },
            "next": ["provisioning"]
        },
        {
            "id": "provisioning",
            "type": "PROVISIONING",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 175, "y": 850 }
            },
            "executor": { "name": "ProvisioningExecutor" },
            "next": ["success"]
        },
        {
            "id": "success",
            "type": "AUTHENTICATION_SUCCESS",
            "properties": {
                "layout": { "width": 180, "height": 90, "x": 175, "y": 1000 }
            },
            "executor": { "name": "AuthAssertExecutor" }
        }
    ]
}
```

## Backend Implementation Changes

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
    Properties   map[string]interface{} `json:"properties,omitempty"`
    InputData    []inputDefinition      `json:"inputData,omitempty"`
    Executor     *executorDefinition    `json:"executor,omitempty"`
    UIDefinition *uiDefinition          `json:"uiDefinition,omitempty"` // NEW
    Next         []string               `json:"next,omitempty"`
}

type uiDefinition struct {
    Layout     *uiLayout     `json:"layout,omitempty"`
    Components []uiComponent `json:"components,omitempty"`
}

type uiLayout struct {
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
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
    TargetNodeID string `json:"targetNodeId"`
}

type executorDefinition struct {
    Name string `json:"name"`
}
```

### 3. Flow Management Service (Simplified)

**File:** `backend/internal/flow/flowmgt/service.go`

**No translation needed!** Service becomes much simpler:

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
        Name:      definition.ID, // Or extract from properties
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

func (v *graphValidator) ValidateGraph(definition *graphDefinition) error {
    // 1. Validate node IDs are unique
    // 2. Validate all "next" references point to valid node IDs
    // 3. Validate flow type-specific requirements:
    //    - AUTHENTICATION: must have AUTHENTICATION_SUCCESS node
    //    - REGISTRATION: must start with REGISTRATION_START, have PROVISIONING, end with AUTH_SUCCESS
    // 4. Validate node-specific requirements:
    //    - UI_INTERACTION: must have uiDefinition
    //    - TASK_EXECUTION: must have executor
    //    - DECISION: must have at least 2 next nodes
    // 5. Validate UI action targets:
    //    - All action targetNodeId must reference valid node IDs
    //    - If UI_INTERACTION has multiple actions, next must be a DECISION node
    //    - DECISION node's next array must include all action target node IDs
    // 6. Validate layout structures exist where required
    // 7. Validate no circular dependencies (optional, can be complex)
    
    return nil
}
```

### 5. Database Schema (No Changes Needed!)

```sql
CREATE TABLE flow_definitions (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    flow_type VARCHAR(50) NOT NULL,
    graph_definition JSONB NOT NULL,  -- Store complete graph as-is
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**Note:** We only store ONE representation now (graph), not two!

### 6. UI Interaction Node Executor

**File:** `backend/internal/flow/core/ui_interaction_node.go` (NEW)

```go
// UIInteractionNode handles UI interaction nodes in the flow
type uiInteractionNode struct {
    baseNode
    uiDefinition *UIDefinition
}

func (n *uiInteractionNode) Execute(ctx *NodeContext) (*common.NodeResponse, error) {
    // Return UI definition to frontend for rendering
    return &common.NodeResponse{
        Status:       common.NodeStatusIncomplete,
        Type:         common.NodeResponseTypeView,
        RequiredData: extractRequiredFields(n.uiDefinition),
        AdditionalData: map[string]string{
            "uiDefinition": marshalUIDefinition(n.uiDefinition),
        },
        Actions:    convertUIActions(n.uiDefinition.Actions),
        NextNodeID: determineNextNode(ctx, n.next),
    }, nil
}
```

## Comparison: Old vs New Approach

| Aspect | Old Approach (Dual Representation) | New Approach (Single Graph) |
|--------|-----------------------------------|----------------------------|
| **Representations** | Step-based (API) + Graph (Internal) | Single graph for both |
| **Translation** | Backend translates steps → graph | No translation needed |
| **Complexity** | High (translator, dual storage) | Low (direct storage) |
| **Frontend Responsibility** | Sends UI steps only | Constructs complete graph |
| **Backend Responsibility** | Translate, validate, store both | Validate, store graph |
| **Consistency** | Risk of translation bugs | Always consistent |
| **Maintenance** | Complex (two formats to maintain) | Simple (one format) |
| **Decision Nodes** | Backend auto-generates | Frontend explicitly adds |
| **Graph Compilation** | Needed during save | Not needed (already a graph) |
| **Runtime Loading** | Load graph from storage | Load graph from storage |
| **API Response** | Return step representation | Return graph representation |

## Benefits of New Approach

1. **Simplicity**: No translation layer, no dual representation
2. **Consistency**: Frontend and backend always in sync (single graph format)
3. **Flexibility**: Frontend has full control over graph structure
4. **Performance**: No compilation overhead during save (store as-is)
5. **Maintainability**: Single source of truth for flow definition
6. **Debugging**: Easier to trace issues (what you see in UI is what's stored in DB)
7. **Extensibility**: Easy to add new node types or properties to graph
8. **Clear Separation**: 
   - Composer layout (properties.layout) for developer UX
   - UI layout (uiDefinition.layout) for end-user UX
9. **Backend Simplification**: Flow management service just validates and stores (no complex translation logic)

## Key Implementation Details

### Validation Logic Examples

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
        // Validate next points to valid node
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
        actionTargets := extractActionTargets(actions)
        if !allTargetsInDecision(actionTargets, nextNode.Next) {
            return errors.New("DECISION node must include all action target nodes")
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
    // ALL nodes must have properties.layout for flow composer positioning
    if node.Properties == nil || node.Properties["layout"] == nil {
        return errors.New("All nodes must have properties.layout for flow composer")
    }
    
    // Validate layout structure
    layout, ok := node.Properties["layout"].(map[string]interface{})
    if !ok {
        return errors.New("properties.layout must be an object")
    }
    
    // Validate required layout fields
    requiredFields := []string{"x", "y", "width", "height"}
    for _, field := range requiredFields {
        if _, exists := layout[field]; !exists {
            return fmt.Errorf("properties.layout missing required field: %s", field)
        }
    }
    
    // UI_INTERACTION nodes should have uiDefinition.layout (optional but recommended)
    if node.Type == "UI_INTERACTION" && node.UIDefinition != nil {
        if node.UIDefinition.Layout == nil {
            // Warning only, not error
            log.Warn("UI_INTERACTION node missing uiDefinition.layout")
        }
    }
    
    return nil
}
```

### Frontend Graph Construction Logic

#### Automatic DECISION Node Insertion

```typescript
function buildGraph(nodes: FlowNode[]): GraphDefinition {
    const graphNodes: NodeDefinition[] = [];
    
    for (const node of nodes) {
        if (node.type === 'UI_INTERACTION') {
            // Extract all ACTION components (recursively)
            const actions = extractActionComponents(node.uiDefinition.components);
            
            if (actions.length > 1) {
                // Multiple actions: insert DECISION node
                const decisionNodeId = `${node.id}_decision`;
                
                // UI node points to decision
                graphNodes.push({
                    ...node,
                    next: [decisionNodeId]
                });
                
                // Create decision node with all action targets
                graphNodes.push({
                    id: decisionNodeId,
                    type: 'DECISION',
                    properties: {
                        layout: generateDecisionLayout(node) // Auto-position
                    },
                    next: actions.map(a => a.action.targetNodeId)
                });
            } else {
                // Single action: direct link
                graphNodes.push({
                    ...node,
                    next: [actions[0].action.targetNodeId]
                });
            }
        } else {
            graphNodes.push(node);
        }
    }
    
    return { nodes: graphNodes };
}

// Helper to recursively extract ACTION components from component tree
function extractActionComponents(components: Component[]): Component[] {
    const actions: Component[] = [];
    
    for (const comp of components) {
        if (comp.category === 'ACTION') {
            actions.push(comp);
        }
        // Recursively check nested components (BLOCK types like FORM)
        if (comp.components && comp.components.length > 0) {
            actions.push(...extractActionComponents(comp.components));
        }
    }
    
    return actions;
}
```

## Finalized Decisions

### 1. Node Type Name ✅

**Decision:** Use `UI_INTERACTION`

### 2. Layout Structure ✅

**Decision:** Two types of layout information:

1. **Node Layout (`properties.layout`)**: Required for ALL nodes
   - Position (x, y) where node is placed in the flow composer canvas
   - Size (width, height) of the node box in the composer
   - User can drag and resize nodes in the flow builder interface
   - Applies to: UI_INTERACTION, TASK_EXECUTION, DECISION, AUTHENTICATION_SUCCESS, REGISTRATION_START, PROVISIONING

2. **UI Layout (`uiDefinition.layout`)**: Only for UI_INTERACTION nodes
   - Size (width, height) of the actual UI dialog/form rendered to end users
   - No position needed (centered or positioned by frontend rendering logic)
   - Does not affect composer visualization

**Example - UI_INTERACTION node (has both):**
```json
{
    "id": "login_view",
    "type": "UI_INTERACTION",
    "properties": {
        "layout": {
            "width": 200,    // Node box size in composer
            "height": 100,
            "x": 100,        // Position in composer canvas
            "y": 200
        }
    },
    "uiDefinition": {
        "layout": {
            "width": 400,    // UI dialog size for end users
            "height": 500
        },
        "components": [...]
    }
}
```

**Example - TASK_EXECUTION node (only properties.layout):**
```json
{
    "id": "basic_auth",
    "type": "TASK_EXECUTION",
    "properties": {
        "layout": {
            "width": 180,
            "height": 90,
            "x": 100,
            "y": 400
        }
    },
    "executor": { "name": "BasicAuthExecutor" }
}
```

### 3. Action Target Validation ✅

**Decision:** Management service validates that:
- All action `targetNodeId` references point to valid nodes
- When UI_INTERACTION has multiple actions, the next node must be a DECISION node
- The DECISION node's `next` array must include all action target node IDs

### 4. Node Types Clarification ✅

**Decision:** No backward compatibility needed. Clean slate with clear node purposes:

- **PROMPT_ONLY**: Gets user input without execution logic (existing, keep as-is)
- **UI_INTERACTION**: Returns UI elements/components to render (new node type)
- **TASK_EXECUTION**: Executes backend logic (existing, keep as-is)
- **DECISION**: Routes flow based on user choice (existing, keep as-is)

These are distinct node types with different purposes. No migration needed - we're building new.

### 5. Success Messages ✅

**Decision:** Do NOT add `uiDefinition` to AUTHENTICATION_SUCCESS.

If success message UI is needed, add a UI_INTERACTION node before AUTHENTICATION_SUCCESS:

```
TASK_EXECUTION (auth) 
    ↓
UI_INTERACTION (success_message) 
    ↓
AUTHENTICATION_SUCCESS
```

Keep node types focused and single-purpose.

## Implementation Roadmap

### Phase 1: Backend Core Changes

1. **Update Constants** (`internal/flow/common/constants.go`)
   - Add `NodeTypeUIInteraction = "UI_INTERACTION"`
   - Keep existing node types as-is

2. **Update Models** (`internal/flow/flowmgt/model.go`)
   - Add `UIDefinition` struct with layout, components, actions
   - Add `UILayout` struct for UI dimensions
   - Add `UIComponent` struct for nested components
   - Add `UIAction` struct with targetNodeId
   - Update `nodeDefinition` to include optional `UIDefinition` field

3. **Implement Validation** (`internal/flow/flowmgt/validator.go`)
   - Validate node IDs unique
   - Validate next references valid
   - Validate action targets valid
   - Validate UI_INTERACTION with multiple actions has DECISION node
   - Validate DECISION includes all action targets
   - Validate flow type-specific requirements

4. **Simplify Service** (`internal/flow/flowmgt/service.go`)
   - Remove translator (no longer needed!)
   - CreateFlow: validate + store graph as-is
   - GetFlow: retrieve and return graph
   - UpdateFlow: validate + update graph
   - ListFlows: paginated list

5. **Database Schema** (`dbscripts/`)
   - Create `flow_definitions` table
   - Single `graph_definition` JSONB column (not two!)
   - Indexes on flow_type, name

6. **Store Layer** (`internal/flow/flowmgt/store.go`)
   - Implement CRUD operations with DBClient
   - Define queries in storeconstants.go

### Phase 2: Flow Execution Support

1. **UI Interaction Node Executor** (`internal/flow/core/ui_interaction_node.go`)
   - Implement node that returns UI definition to frontend
   - Extract required fields from components
   - Convert UI actions to node actions
   - Return NodeResponse with VIEW type

2. **Register Node Type** (`internal/flow/core/factory.go`)
   - Add UI_INTERACTION case in node creation
   - Wire up to uiInteractionNode implementation

3. **Update Node Response** (`internal/flow/common/model.go`)
   - Ensure NodeResponse can carry UI definition data
   - Add any missing fields for UI metadata

### Phase 3: API Layer

1. **HTTP Handlers** (`internal/flow/flowmgt/handler.go`)
   - Implement CREATE, GET, UPDATE, DELETE, LIST handlers
   - Parse request bodies (graph JSON)
   - Return appropriate responses and errors

2. **Route Registration** (`internal/flow/flowmgt/init.go`)
   - Register /flows endpoints
   - Wire up dependencies (store, validator, service, handler)
   - Apply CORS middleware

3. **Error Constants** (`internal/flow/flowmgt/errorconstants.go`)
   - Define FMS-xxxx error codes
   - Define error messages

4. **Update Service Manager** (`cmd/server/servicemanager.go`)
   - Call flowmgt.Initialize(mux)

### Phase 4: Frontend Support

1. **Graph Builder Utilities**
   - Helper functions to construct graph JSON
   - Automatic DECISION node insertion logic
   - Validation helpers

2. **Flow Composer UI**
   - Drag-and-drop canvas for nodes
   - Node property editors
   - UI component builder for UI_INTERACTION nodes
   - Action button configuration
   - Graph validation and preview

### Phase 5: Testing

1. **Unit Tests**
   - Validator tests (all validation rules)
   - Service tests (mock store)
   - Store tests (go-sqlmock)
   - Node executor tests

2. **Integration Tests** (`tests/integration/flowmgt/`)
   - Create flow via API
   - Update flow via API
   - List flows with pagination
   - Execute flow with UI_INTERACTION nodes
   - Validate error responses

### Phase 6: Documentation & Examples

1. **Update OpenAPI Spec** (`api/flow-management.yaml`)
   - Define graph structure with UI_INTERACTION nodes
   - Update examples to show new format
   - Remove step-based examples

2. **Example Flows**
   - Create example graph JSONs for common scenarios
   - Basic auth
   - Social login
   - Multi-option flows
   - Registration flows

3. **Developer Guide**
   - How to construct graphs
   - UI_INTERACTION node guide
   - Action and DECISION node patterns
   - Validation rules documentation

## Summary of Key Decisions

| Decision Point | Finalized Approach |
|----------------|-------------------|
| **Node Type Name** | `UI_INTERACTION` |
| **Composer Layout** | `properties.layout` (x, y, width, height) - required for ALL nodes |
| **UI Layout** | `uiDefinition.layout` (width, height) - only for UI_INTERACTION nodes |
| **Action Validation** | Management service validates action targets and DECISION node correlation |
| **Backward Compatibility** | Not needed - clean implementation of distinct node types |
| **Success Messages** | Use separate UI_INTERACTION node before AUTHENTICATION_SUCCESS |
| **Translation Layer** | None - frontend sends graph, backend stores as-is |
| **Storage Model** | Single graph representation in database |
| **Frontend Responsibility** | Construct complete graph including DECISION node insertion |
| **Backend Responsibility** | Validate graph structure and store/retrieve |

## References

- Original implementation guide: `/docs/contributing/flow-management-implementation.md` (obsolete, replaced by this document)
- OpenAPI Specification: `/api/WIP/flow-management.yaml` (needs update for graph format)
- Existing graph examples: `/backend/cmd/server/repository/resources/graphs/`
- Flow engine code: `/backend/internal/flow/`
- Flow models: `/backend/internal/flow/flowmgt/model.go`
