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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	appmodel "github.com/asgardeo/thunder/internal/application/model"
	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/tests/mocks/flow/coremock"
	"github.com/asgardeo/thunder/tests/mocks/userschemamock"
)

type UserTypeResolverTestSuite struct {
	suite.Suite
	mockUserSchemaService *userschemamock.UserSchemaServiceInterfaceMock
	mockFlowFactory       *coremock.FlowFactoryInterfaceMock
	executor              *userTypeResolver
}

func TestUserTypeResolverSuite(t *testing.T) {
	suite.Run(t, new(UserTypeResolverTestSuite))
}

func (suite *UserTypeResolverTestSuite) SetupTest() {
	suite.mockUserSchemaService = userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())
	suite.mockFlowFactory = coremock.NewFlowFactoryInterfaceMock(suite.T())

	// Mock the CreateExecutor method to return a base executor
	suite.mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, flowcm.ExecutorTypeRegistration,
		[]flowcm.InputData{}, []flowcm.InputData{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	suite.executor = newUserTypeResolver(suite.mockFlowFactory, suite.mockUserSchemaService)
}

func createMockUserTypeResolverExecutor(t *testing.T) flowcore.ExecutorInterface {
	mockExec := coremock.NewExecutorInterfaceMock(t)
	mockExec.On("GetName").Return(ExecutorNameUserTypeResolver).Maybe()
	mockExec.On("GetType").Return(flowcm.ExecutorTypeRegistration).Maybe()
	mockExec.On("GetDefaultExecutorInputs").Return([]flowcm.InputData{}).Maybe()
	mockExec.On("GetPrerequisites").Return([]flowcm.InputData{}).Maybe()
	return mockExec
}

func (suite *UserTypeResolverTestSuite) TestNewUserTypeResolver() {
	mockFlowFactory := coremock.NewFlowFactoryInterfaceMock(suite.T())
	mockUserSchemaService := userschemamock.NewUserSchemaServiceInterfaceMock(suite.T())

	mockFlowFactory.On("CreateExecutor", ExecutorNameUserTypeResolver, flowcm.ExecutorTypeRegistration,
		[]flowcm.InputData{}, []flowcm.InputData{}).
		Return(createMockUserTypeResolverExecutor(suite.T()))

	executor := newUserTypeResolver(mockFlowFactory, mockUserSchemaService)

	assert.NotNil(suite.T(), executor)
	assert.Equal(suite.T(), ExecutorNameUserTypeResolver, executor.GetName())
}

func (suite *UserTypeResolverTestSuite) TestExecute_NonRegistrationFlow() {
	testCases := []struct {
		name     string
		flowType flowcm.FlowType
	}{
		{
			name:     "Authentication flow",
			flowType: flowcm.FlowTypeAuthentication,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:   "flow-123",
				FlowType: tc.flowType,
				Application: appmodel.ApplicationProcessedDTO{
					AllowedUserTypes: []string{"employee"},
				},
			}

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecComplete, result.Status)
			assert.Empty(suite.T(), result.RuntimeData[inputUserType])
			suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetOUForUserType")
		})
	}
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_Success() {
	testCases := []struct {
		name             string
		allowedUserTypes []string
		providedUserType string
		expectedOUID     string
	}{
		{
			name:             "Valid user type with OU",
			allowedUserTypes: []string{"employee", "customer"},
			providedUserType: "employee",
			expectedOUID:     "ou-123",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()

			ctx := &flowcore.NodeContext{
				FlowID:   "flow-123",
				FlowType: flowcm.FlowTypeRegistration,
				Application: appmodel.ApplicationProcessedDTO{
					AllowedUserTypes: tc.allowedUserTypes,
				},
				UserInputData: map[string]string{
					inputUserType: tc.providedUserType,
				},
				RuntimeData: map[string]string{},
			}

			ouID := tc.expectedOUID
			suite.mockUserSchemaService.On("GetOUForUserType", tc.providedUserType).
				Return(&ouID, nil)

			result, err := suite.executor.Execute(ctx)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), flowcm.ExecComplete, result.Status)
			assert.Equal(suite.T(), tc.providedUserType, result.RuntimeData[inputUserType])
			assert.Equal(suite.T(), tc.expectedOUID, result.RuntimeData[defaultOUIDKey])

			suite.mockUserSchemaService.AssertExpectations(suite.T())
		})
	}
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_NoOU() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputData: map[string]string{
			inputUserType: "employee",
		},
		RuntimeData: map[string]string{},
	}

	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_NotAllowed() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputData: map[string]string{
			inputUserType: "partner",
		},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Registration not allowed for user type: partner", result.FailureReason)
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetOUForUserType")
}

func (suite *UserTypeResolverTestSuite) TestExecute_UserTypeProvidedInInput_OUResolutionFails() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{
			inputUserType: "employee",
		},
		RuntimeData: map[string]string{},
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve OU",
	}
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve organization unit")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_NoAllowedUserTypes() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{},
		},
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecFailure, result.Status)
	assert.Equal(suite.T(), "Self-registration not available for this application", result.FailureReason)
	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetOUForUserType")
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_Success() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{},
	}

	ouID := "ou-123"
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(&ouID, nil)

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecComplete, result.Status)
	assert.Equal(suite.T(), "employee", result.RuntimeData[inputUserType])
	assert.Equal(suite.T(), "ou-123", result.RuntimeData[defaultOUIDKey])

	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_NoOU() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{},
	}

	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, nil)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "no organization unit found")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_SingleAllowedUserType_OUResolutionFails() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee"},
		},
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{},
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve OU",
	}
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, svcErr)

	result, err := suite.executor.Execute(ctx)

	assert.Error(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Contains(suite.T(), err.Error(), "failed to resolve organization unit")
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestExecute_MultipleAllowedUserTypes_PromptUser() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee", "customer", "partner"},
		},
		UserInputData: map[string]string{},
		RuntimeData:   map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.RequiredData)
	assert.Len(suite.T(), result.RequiredData, 1)

	requiredInput := result.RequiredData[0]
	assert.Equal(suite.T(), inputUserType, requiredInput.Name)
	assert.Equal(suite.T(), "dropdown", requiredInput.Type)
	assert.True(suite.T(), requiredInput.Required)
	assert.ElementsMatch(suite.T(), []string{"employee", "customer", "partner"}, requiredInput.Options)

	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetOUForUserType")
}

func (suite *UserTypeResolverTestSuite) TestExecute_EmptyUserTypeInput() {
	suite.SetupTest()

	ctx := &flowcore.NodeContext{
		FlowID:   "flow-123",
		FlowType: flowcm.FlowTypeRegistration,
		Application: appmodel.ApplicationProcessedDTO{
			AllowedUserTypes: []string{"employee", "customer"},
		},
		UserInputData: map[string]string{
			inputUserType: "",
		},
		RuntimeData: map[string]string{},
	}

	result, err := suite.executor.Execute(ctx)

	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), flowcm.ExecUserInputRequired, result.Status)
	assert.NotEmpty(suite.T(), result.RequiredData)
	assert.Len(suite.T(), result.RequiredData, 1)

	requiredInput := result.RequiredData[0]
	assert.Equal(suite.T(), inputUserType, requiredInput.Name)
	assert.Equal(suite.T(), "dropdown", requiredInput.Type)

	suite.mockUserSchemaService.AssertNotCalled(suite.T(), "GetOUForUserType")
}

func (suite *UserTypeResolverTestSuite) TestResolveAndSetDefaultOUID_Success() {
	suite.SetupTest()

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	ouID := "ou-123"
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(&ouID, nil)

	err := suite.executor.resolveAndSetDefaultOUID("employee", execResp)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "ou-123", execResp.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestResolveAndSetDefaultOUID_NoOUFound() {
	suite.SetupTest()

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, nil)

	err := suite.executor.resolveAndSetDefaultOUID("employee", execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no organization unit found")
	assert.Empty(suite.T(), execResp.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestResolveAndSetDefaultOUID_EmptyOUIDPointer() {
	suite.SetupTest()

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	emptyOUID := ""
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(&emptyOUID, nil)

	err := suite.executor.resolveAndSetDefaultOUID("employee", execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "no organization unit found")
	assert.Empty(suite.T(), execResp.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}

func (suite *UserTypeResolverTestSuite) TestResolveAndSetDefaultOUID_ServiceError() {
	suite.SetupTest()

	execResp := &flowcm.ExecutorResponse{
		RuntimeData: make(map[string]string),
	}

	svcErr := &serviceerror.ServiceError{
		Type:             serviceerror.ServerErrorType,
		Code:             "SCHEMA-500",
		Error:            "Internal Server Error",
		ErrorDescription: "Failed to retrieve OU",
	}
	suite.mockUserSchemaService.On("GetOUForUserType", "employee").
		Return(nil, svcErr)

	err := suite.executor.resolveAndSetDefaultOUID("employee", execResp)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to resolve organization unit")
	assert.Empty(suite.T(), execResp.RuntimeData[defaultOUIDKey])
	suite.mockUserSchemaService.AssertExpectations(suite.T())
}
