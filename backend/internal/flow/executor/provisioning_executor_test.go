/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package executor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authncm "github.com/asgardeo/thunder/internal/authn/common"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/user"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/usermock"
)

type ProvisioningExecutorTestSuite struct {
	suite.Suite
	mockUserService *usermock.UserServiceInterfaceMock
	mockFlowFactory *coremock.FlowFactoryInterfaceMock
	executor        *provisioningExecutor
}

func TestProvisioningExecutorSuite(t *testing.T) {
	suite.Run(t, new(ProvisioningExecutorTestSuite))
}

func (suite *ProvisioningExecutorTestSuite) SetupTest() {
	suite.mockUserService = usermock.NewUserServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	// Mock the embedded identifying executor first
	identifyingMock := suite.createMockIdentifyingExecutor()
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameIdentifying, flowcm.ExecutorTypeUtility,
		mock.Anything, mock.Anything).Return(identifyingMock).Maybe()

	mockExec := suite.createMockProvisioningExecutor()
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameProvisioning, flowcm.ExecutorTypeRegistration,
		[]flowcm.InputData{}, []flowcm.InputData{}).Return(mockExec)

	suite.executor = newProvisioningExecutor(suite.mockFlowFactory, suite.mockUserService)
}

func (suite *ProvisioningExecutorTestSuite) createMockIdentifyingExecutor() flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameIdentifying).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeUtility).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	return mockExec
}

func (suite *ProvisioningExecutorTestSuite) createMockProvisioningExecutor() flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(suite.T())
	mockExec.On("GetName").Return(ExecutorNameProvisioning).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeRegistration).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("CheckInputData", mock.Anything, mock.Anything).Return(
		func(ctx *flowcore.NodeContext, execResp *flowcm.ExecutorResponse) bool {
			if len(ctx.NodeInputData) == 0 {
				return false
			}
			for _, input := range ctx.NodeInputData {
				if _, ok := ctx.UserInputData[input.Name]; !ok {
					if _, ok := ctx.RuntimeData[input.Name]; !ok {
						execResp.RequiredData = append(execResp.RequiredData, input)
					}
				}
			}
			return len(execResp.RequiredData) > 0
		}).Maybe()
	mockExec.On("GetRequiredData", mock.Anything).Return([]flowcm.InputData{}).Maybe()
	return mockExec
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_NonRegistrationFlow() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeAuthentication,
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_Success() {
	attrs := map[string]interface{}{"username": "newuser", "email": "new@example.com"}
	attrsJSON, _ := json.Marshal(attrs)

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"username": "newuser",
			"email":    "new@example.com",
		},
		RuntimeData: map[string]string{
			"ouId":     "ou-123",
			"userType": "INTERNAL",
		},
		NodeInputData: []flowcm.InputData{
			{Name: "username", Type: "string", Required: true},
			{Name: "email", Type: "string", Required: true},
		},
	}

	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		"username": "newuser",
		"email":    "new@example.com",
	}).Return(nil, &user.ErrorUserNotFound)

	createdUser := &user.User{
		ID:               "user-new",
		OrganizationUnit: "ou-123",
		Type:             "INTERNAL",
		Attributes:       attrsJSON,
	}

	suite.mockUserService.On("CreateUser", mock.MatchedBy(func(u *user.User) bool {
		return u.OrganizationUnit == "ou-123" && u.Type == "INTERNAL"
	})).Return(createdUser, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecComplete, resp.Status)
	assert.True(suite.T(), resp.AuthenticatedUser.IsAuthenticated)
	assert.Equal(suite.T(), "user-new", resp.AuthenticatedUser.UserID)
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_UserAlreadyExists() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"username": "existinguser",
		},
		NodeInputData: []flowcm.InputData{{Name: "username", Type: "string", Required: true}},
	}

	userID := "user-existing"
	suite.mockUserService.On("IdentifyUser", map[string]interface{}{
		"username": "existinguser",
	}).Return(&userID, nil)

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "User already exists")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_NoUserAttributes() {
	ctx := &flowcore.NodeContext{
		FlowID:        "flow-123",
		FlowType:      flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{},
		NodeInputData: []flowcm.InputData{},
	}

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "No user attributes provided")
}

func (suite *ProvisioningExecutorTestSuite) TestExecute_CreateUserFails() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		UserInputData: map[string]string{
			"username": "newuser",
		},
		RuntimeData: map[string]string{
			"ouId":     "ou-123",
			"userType": "INTERNAL",
		},
		NodeInputData: []flowcm.InputData{{Name: "username", Type: "string", Required: true}},
	}

	suite.mockUserService.On("IdentifyUser", mock.Anything).Return(nil, &user.ErrorUserNotFound)
	suite.mockUserService.On("CreateUser", mock.Anything).
		Return(nil, &serviceerror.ServiceError{Error: "creation failed"})

	resp, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), resp)
	assert.Equal(suite.T(), flowcm.ExecFailure, resp.Status)
	assert.Contains(suite.T(), resp.FailureReason, "Failed to create user")
	suite.mockUserService.AssertExpectations(suite.T())
}

func (suite *ProvisioningExecutorTestSuite) TestCheckInputData_AttributesFromAuthUser() {
	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		AuthenticatedUser: authncm.AuthenticatedUser{
			Attributes: map[string]interface{}{"email": "test@example.com"},
		},
		NodeInputData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
	}

	execResp := &flowcm.ExecutorResponse{
		RequiredData: []flowcm.InputData{{Name: "email", Type: "string", Required: true}},
		RuntimeData:  make(map[string]string),
	}

	result := suite.executor.CheckInputData(ctx, execResp)

	assert.False(suite.T(), result)
	assert.Empty(suite.T(), execResp.RequiredData)
	assert.Equal(suite.T(), "test@example.com", execResp.RuntimeData["email"])
}

func (suite *ProvisioningExecutorTestSuite) TestGetInputAttributes_FromMultipleSources() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{"username": "testuser", "code": "auth-code"},
		RuntimeData:   map[string]string{"email": "test@example.com"},
		NodeInputData: []flowcm.InputData{},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Contains(suite.T(), result, "username")
	assert.Contains(suite.T(), result, "email")
	assert.NotContains(suite.T(), result, "code")
}

func (suite *ProvisioningExecutorTestSuite) TestGetInputAttributes_FilterNonUserAttributes() {
	ctx := &flowcore.NodeContext{
		UserInputData: map[string]string{
			"username": "testuser",
			"userID":   "user-123",
			"code":     "auth-code",
			"nonce":    "test-nonce",
		},
		NodeInputData: []flowcm.InputData{},
	}

	result := suite.executor.getInputAttributes(ctx)

	assert.Contains(suite.T(), result, "username")
	assert.NotContains(suite.T(), result, "userID")
	assert.NotContains(suite.T(), result, "code")
	assert.NotContains(suite.T(), result, "nonce")
}
