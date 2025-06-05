/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
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

package engine

import (
	"testing"

	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Mock implementations for testing
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx *model.NodeContext) (*model.ExecutorResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ExecutorResponse), args.Error(1)
}

func (m *MockExecutor) GetID() string {
	return m.Called().String(0)
}

func (m *MockExecutor) GetName() string {
	return m.Called().String(0)
}

func (m *MockExecutor) GetProperties() model.ExecutorProperties {
	return m.Called().Get(0).(model.ExecutorProperties)
}

func (m *MockExecutor) GetDefaultExecutorInputs() []model.InputData {
	return m.Called().Get(0).([]model.InputData)
}

func (m *MockExecutor) CheckInputData(ctx *model.NodeContext, execResp *model.ExecutorResponse) bool {
	return m.Called(ctx, execResp).Bool(0)
}

type MockNode struct {
	mock.Mock
}

func (m *MockNode) Execute(ctx *model.NodeContext) (*model.NodeResponse, *serviceerror.ServiceError) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*serviceerror.ServiceError)
	}
	return args.Get(0).(*model.NodeResponse), nil
}

func (m *MockNode) GetID() string {
	return m.Called().String(0)
}

func (m *MockNode) GetType() string {
	return m.Called().String(0)
}

func (m *MockNode) IsStartNode() bool {
	return m.Called().Bool(0)
}

func (m *MockNode) SetAsStartNode(isStart bool) {
	m.Called(isStart)
}

func (m *MockNode) IsFinalNode() bool {
	return m.Called().Bool(0)
}

func (m *MockNode) SetAsFinalNode(isFinal bool) {
	m.Called(isFinal)
}

func (m *MockNode) GetNextNodeID() string {
	return m.Called().String(0)
}

func (m *MockNode) SetNextNodeID(nextNodeID string) {
	m.Called(nextNodeID)
}

func (m *MockNode) GetPreviousNodeID() string {
	return m.Called().String(0)
}

func (m *MockNode) SetPreviousNodeID(previousNodeID string) {
	m.Called(previousNodeID)
}

func (m *MockNode) GetInputData() []model.InputData {
	return m.Called().Get(0).([]model.InputData)
}

func (m *MockNode) SetInputData(inputData []model.InputData) {
	m.Called(inputData)
}

func (m *MockNode) GetExecutorConfig() *model.ExecutorConfig {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*model.ExecutorConfig)
}

func (m *MockNode) SetExecutorConfig(executorConfig *model.ExecutorConfig) {
	m.Called(executorConfig)
}

func (m *MockNode) GetExecutor() model.ExecutorInterface {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(model.ExecutorInterface)
}

func (m *MockNode) SetExecutor(executor model.ExecutorInterface) {
	m.Called(executor)
}

type MockGraph struct {
	mock.Mock
}

func (m *MockGraph) GetID() string {
	return m.Called().String(0)
}

func (m *MockGraph) GetStartNodeID() string {
	return m.Called().String(0)
}

func (m *MockGraph) GetNode(nodeID string) (model.NodeInterface, bool) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(model.NodeInterface), args.Bool(1)
}

func (m *MockGraph) AddEdge(sourceNodeID, targetNodeID string) error {
	args := m.Called(sourceNodeID, targetNodeID)
	return args.Error(0)
}

func (m *MockGraph) AddNode(node model.NodeInterface) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockGraph) GetNodes() map[string]model.NodeInterface {
	args := m.Called()
	return args.Get(0).(map[string]model.NodeInterface)
}

func (m *MockGraph) GetEdges() map[string][]string {
	args := m.Called()
	return args.Get(0).(map[string][]string)
}

func (m *MockGraph) SetStartNodeID(startNodeID string) error {
	args := m.Called(startNodeID)
	return args.Error(0)
}

func (m *MockGraph) ToJSON() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// FlowEngineTestSuite defines the test suite for the flow engine
type FlowEngineTestSuite struct {
	suite.Suite
	engine FlowEngineInterface
}

// SetupTest initializes objects for each test
func (suite *FlowEngineTestSuite) SetupTest() {
	suite.engine = GetFlowEngine()
}

// TestSuccessfulFlowExecution tests a successful flow execution with a complete node
func (suite *FlowEngineTestSuite) TestSuccessfulFlowExecution() {
	// Mock setup
	mockNode := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Define behavior
	mockNode.On("GetID").Return("node1")
	mockNode.On("GetType").Return("test")
	mockNode.On("IsStartNode").Return(true)
	mockNode.On("IsFinalNode").Return(false)

	executorConfig := &model.ExecutorConfig{Name: "testExecutor"}
	mockNode.On("GetExecutorConfig").Return(executorConfig)
	mockNode.On("GetNextNodeID").Return("")
	mockNode.On("GetPreviousNodeID").Return("")
	mockNode.On("GetInputData").Return([]model.InputData{})
	mockNode.On("GetExecutor").Return(mockExecutor)
	mockNode.On("SetExecutor", mock.Anything).Return()

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode, true)

	nodeResp := &model.NodeResponse{
		Status:    constants.NodeStatusComplete,
		Assertion: "test-assertion",
	}
	mockNode.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		AppID:  "app123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return an error")
	assert.Equal(suite.T(), "flow123", flowStep.FlowID, "Flow ID should match")
	assert.Equal(suite.T(), constants.FlowStatusComplete, flowStep.Status, "Flow status should be complete")
	assert.Equal(suite.T(), "test-assertion", flowStep.Assertion, "Assertion should be present")

	mockNode.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
	mockExecutor.AssertExpectations(suite.T())
}

// TestFlowWithMissingGraph tests executing a flow with a missing graph
func (suite *FlowEngineTestSuite) TestFlowWithMissingGraph() {
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  nil, // Missing graph
	}

	flowStep, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err, "Should return an error")
	assert.Equal(suite.T(), constants.ErrorFlowGraphNotInitialized.Code, err.Code, "Error code should match")
	assert.Equal(suite.T(), "flow123", flowStep.FlowID, "Flow ID should match")
}

// TestFlowWithMissingStartNode tests executing a flow with a missing start node
func (suite *FlowEngineTestSuite) TestFlowWithMissingStartNode() {
	mockGraph := new(MockGraph)
	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(nil, false) // Start node not found

	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	assert.NotNil(suite.T(), err, "Should return an error")
	assert.Equal(suite.T(), constants.ErrorStartNodeNotFoundInGraph.Code, err.Code, "Error code should match")
	assert.Equal(suite.T(), "flow123", flowStep.FlowID, "Flow ID should match")

	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithMultipleNodes tests executing a flow with multiple nodes
func (suite *FlowEngineTestSuite) TestFlowWithMultipleNodes() {
	// Setup mocks
	mockNode1 := new(MockNode)
	mockNode2 := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor1 := new(MockExecutor)
	mockExecutor2 := new(MockExecutor)

	// Node 1 configuration
	mockNode1.On("GetID").Return("node1")
	mockNode1.On("GetType").Return("test1")
	mockNode1.On("IsStartNode").Return(true)
	mockNode1.On("IsFinalNode").Return(false)

	executorConfig1 := &model.ExecutorConfig{Name: "testExecutor1"}
	mockNode1.On("GetExecutorConfig").Return(executorConfig1)
	mockNode1.On("GetNextNodeID").Return("node2")
	mockNode1.On("GetPreviousNodeID").Return("")
	mockNode1.On("GetInputData").Return([]model.InputData{})
	mockNode1.On("GetExecutor").Return(mockExecutor1)
	mockNode1.On("SetExecutor", mock.Anything).Return()

	// Node 2 configuration
	mockNode2.On("GetID").Return("node2")
	mockNode2.On("GetType").Return("test2")
	mockNode2.On("IsStartNode").Return(false)
	mockNode2.On("IsFinalNode").Return(true)

	executorConfig2 := &model.ExecutorConfig{Name: "testExecutor2"}
	mockNode2.On("GetExecutorConfig").Return(executorConfig2)
	mockNode2.On("GetNextNodeID").Return("") // No next node (end of flow)
	mockNode2.On("GetPreviousNodeID").Return("node1")
	mockNode2.On("GetInputData").Return([]model.InputData{})
	mockNode2.On("GetExecutor").Return(mockExecutor2)
	mockNode2.On("SetExecutor", mock.Anything).Return()

	// Graph configuration
	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode1, true)
	mockGraph.On("GetNode", "node2").Return(mockNode2, true)

	// Node responses
	nodeResp1 := &model.NodeResponse{
		Status: constants.NodeStatusComplete,
	}
	nodeResp2 := &model.NodeResponse{
		Status:    constants.NodeStatusComplete,
		Assertion: "final-assertion",
	}

	mockNode1.On("Execute", mock.Anything).Return(nodeResp1, nil)
	mockNode2.On("Execute", mock.Anything).Return(nodeResp2, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return an error")
	assert.Equal(suite.T(), constants.FlowStatusComplete, flowStep.Status, "Flow status should be complete")
	assert.Equal(suite.T(), "final-assertion", flowStep.Assertion, "Assertion should be present")

	mockNode1.AssertExpectations(suite.T())
	mockNode2.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithIncompleteRedirection tests a flow that requires redirection
func (suite *FlowEngineTestSuite) TestFlowWithIncompleteRedirection() {
	// Setup mocks
	mockNode := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Node configuration
	mockNode.On("GetID").Return("node1")
	mockNode.On("GetType").Return("test")
	mockNode.On("GetExecutorConfig").Return(model.ExecutorConfig{Name: "testExecutor"})
	mockNode.On("GetInputData").Return(map[string]string{"key": "value"})
	mockNode.On("GetExecutor").Return(mockExecutor)
	mockNode.On("SetExecutor", mock.Anything).Return()

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode, true)

	// Node response with redirection
	nodeResp := &model.NodeResponse{
		Status: constants.NodeStatusIncomplete,
		Type:   constants.NodeResponseTypeRedirection,
		AdditionalInfo: map[string]string{
			constants.DataRedirectURL: "https://example.com/auth",
		},
	}

	mockNode.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return an error")
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, flowStep.Status, "Flow status should be incomplete")
	assert.Equal(suite.T(), constants.StepTypeRedirection, flowStep.Type, "Flow type should be redirection")
	assert.Equal(suite.T(), "https://example.com/auth", flowStep.AdditionalInfo[constants.DataRedirectURL],
		"Redirect URL should be present")

	mockNode.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithIncompleteView tests a flow that requires user input
func (suite *FlowEngineTestSuite) TestFlowWithIncompleteView() {
	// Setup mocks
	mockNode := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Node configuration
	mockNode.On("GetID").Return("node1")
	mockNode.On("GetType").Return("test")
	mockNode.On("GetExecutorConfig").Return(model.ExecutorConfig{Name: "testExecutor"})
	mockNode.On("GetInputData").Return(map[string]string{"key": "value"})
	mockNode.On("GetExecutor").Return(mockExecutor)
	mockNode.On("SetExecutor", mock.Anything).Return()

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode, true)

	// Node response requiring user input
	nodeResp := &model.NodeResponse{
		Status: constants.NodeStatusIncomplete,
		Type:   constants.NodeResponseTypeView,
		RequiredData: []model.InputData{
			{Name: "username", Type: "text", Required: true},
			{Name: "password", Type: "password", Required: true},
		},
	}

	mockNode.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return an error")
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, flowStep.Status, "Flow status should be incomplete")
	assert.Equal(suite.T(), constants.StepTypeView, flowStep.Type, "Flow type should be view")
	assert.Equal(suite.T(), 2, len(flowStep.InputData), "Should have 2 input fields")
	assert.Equal(suite.T(), "username", flowStep.InputData[0].Name, "First input should be username")
	assert.Equal(suite.T(), "password", flowStep.InputData[1].Name, "Second input should be password")

	mockNode.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithError tests a flow that encounters an error
func (suite *FlowEngineTestSuite) TestFlowWithError() {
	// Setup mocks
	mockNode := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Node configuration
	mockNode.On("GetID").Return("node1")
	mockNode.On("GetType").Return("test")
	mockNode.On("GetExecutorConfig").Return(model.ExecutorConfig{Name: "testExecutor"})
	mockNode.On("GetInputData").Return(map[string]string{"key": "value"})
	mockNode.On("GetExecutor").Return(mockExecutor)
	mockNode.On("SetExecutor", mock.Anything).Return()

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode, true)

	// Node response with error
	nodeResp := &model.NodeResponse{
		Status:        constants.NodeStatusFailure,
		FailureReason: "Authentication failed: Invalid credentials",
	}

	mockNode.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return service error")
	assert.Equal(suite.T(), constants.FlowStatusError, flowStep.Status, "Flow status should be error")
	assert.Equal(suite.T(), "Authentication failed: Invalid credentials", flowStep.FailureReason,
		"Failure reason should be present")

	mockNode.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithUnsupportedNodeStatus tests handling of an unsupported node status
func (suite *FlowEngineTestSuite) TestFlowWithUnsupportedNodeStatus() {
	// Setup mocks
	mockNode := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Node configuration
	mockNode.On("GetID").Return("node1")
	mockNode.On("GetType").Return("test")
	mockNode.On("GetExecutorConfig").Return(model.ExecutorConfig{Name: "testExecutor"})
	mockNode.On("GetInputData").Return(map[string]string{"key": "value"})
	mockNode.On("GetExecutor").Return(mockExecutor)
	mockNode.On("SetExecutor", mock.Anything).Return()

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode, true)

	// Node response with unsupported status
	nodeResp := &model.NodeResponse{
		Status: "INVALID_STATUS",
	}

	mockNode.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	_, err := suite.engine.Execute(ctx)

	// Assertions
	assert.NotNil(suite.T(), err, "Should return an error")
	assert.Equal(suite.T(), constants.ErrorUnsupportedNodeResponseStatus.Code, err.Code,
		"Error code should match")

	mockNode.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestFlowWithPromptOnlyStatus tests a flow with a prompt-only node status
func (suite *FlowEngineTestSuite) TestFlowWithPromptOnlyStatus() {
	// Setup mocks
	mockNode1 := new(MockNode)
	mockNode2 := new(MockNode)
	mockGraph := new(MockGraph)
	mockExecutor := new(MockExecutor)

	// Node configuration
	mockNode1.On("GetID").Return("node1")
	mockNode1.On("GetType").Return("test")
	mockNode1.On("GetExecutorConfig").Return(model.ExecutorConfig{Name: "testExecutor"})
	mockNode1.On("GetInputData").Return(map[string]string{"key": "value"})
	mockNode1.On("GetExecutor").Return(mockExecutor)
	mockNode1.On("SetExecutor", mock.Anything).Return()
	mockNode1.On("GetNextNodeID").Return("node2")

	mockNode2.On("GetID").Return("node2")

	mockGraph.On("GetStartNodeID").Return("node1")
	mockGraph.On("GetNode", "node1").Return(mockNode1, true)
	mockGraph.On("GetNode", "node2").Return(mockNode2, true)

	// Node response with prompt-only status
	nodeResp := &model.NodeResponse{
		Status: constants.NodeStatusPromptOnly,
		Type:   constants.NodeResponseTypeView,
		RequiredData: []model.InputData{
			{Name: "otp", Type: "text", Required: true},
		},
	}

	mockNode1.On("Execute", mock.Anything).Return(nodeResp, nil)

	// Setup context and execute
	ctx := &model.EngineContext{
		FlowID: "flow123",
		Graph:  mockGraph,
	}

	flowStep, err := suite.engine.Execute(ctx)

	// Assertions
	assert.Nil(suite.T(), err, "Should not return an error")
	assert.Equal(suite.T(), constants.FlowStatusIncomplete, flowStep.Status, "Flow status should be incomplete")
	assert.Equal(suite.T(), constants.StepTypeView, flowStep.Type, "Flow type should be view")
	assert.Equal(suite.T(), 1, len(flowStep.InputData), "Should have 1 input field")
	assert.Equal(suite.T(), "otp", flowStep.InputData[0].Name, "Input should be otp")

	mockNode1.AssertExpectations(suite.T())
	mockNode2.AssertExpectations(suite.T())
	mockGraph.AssertExpectations(suite.T())
}

// TestMainFlowEngineSuite runs the test suite
func TestFlowEngineSuite(t *testing.T) {
	suite.Run(t, new(FlowEngineTestSuite))
}
