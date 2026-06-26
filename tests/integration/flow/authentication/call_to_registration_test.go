/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package authentication

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/flow/common"
	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

// Tests an AUTHENTICATION flow that CALLs a REGISTRATION sub-flow. The sub-flow
// provisions a new user (which authenticates them); on return, the caller's
// AuthAssertExecutor issues a JWT assertion.

var (
	callRegOU = testutils.OrganizationUnit{
		Handle:      "call_to_reg_ou",
		Name:        "Test OU for CALL→Registration",
		Description: "Organization unit for call-to-registration test",
	}

	callRegUserType = testutils.UserType{
		Name: "call_to_reg_user",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type":       "string",
				"credential": true,
			},
			"email": map[string]interface{}{
				"type": "string",
			},
		},
		AllowSelfRegistration: true,
	}

	// Registration sub-flow: prompts for credentials, then provisions.
	callRegSubFlow = testutils.Flow{
		Name:     "Call-To-Reg Registration Sub-Flow",
		FlowType: "REGISTRATION",
		Handle:   "call_to_reg_sub_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_reg",
			},
			{
				"id":   "prompt_reg",
				"type": "PROMPT",
				"prompts": []map[string]interface{}{
					{
						"inputs": []map[string]interface{}{
							{
								"ref":        "input_username",
								"identifier": "username",
								"type":       "TEXT_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_password",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
							{
								"ref":        "input_email",
								"identifier": "email",
								"type":       "EMAIL_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_register",
							"nextNode": "provisioning",
						},
					},
				},
			},
			{
				"id":   "provisioning",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "ProvisioningExecutor",
				},
				"onSuccess":    "end",
				"onIncomplete": "prompt_reg",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	// Caller authentication flow: START → CALL(reg sub-flow) → auth_assert → END.
	callRegAuthFlow = testutils.Flow{
		Name:     "Call-To-Reg Auth Caller Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "call_to_reg_auth_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "call_reg",
			},
			{
				"id":        "call_reg",
				"type":      "CALL",
				"flow":      map[string]interface{}{"ref": ""}, // populated in SetupSuite
				"onSuccess": "auth_assert",
			},
			{
				"id":   "auth_assert",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "AuthAssertExecutor",
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	callRegTestApp = testutils.Application{
		Name:                      "Call-To-Reg Test App",
		Description:               "App for CALL→Registration test",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "call_to_reg_client",
		ClientSecret:              "secret123",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{"call_to_reg_user"},
		AssertionConfig: map[string]interface{}{
			"userAttributes": []string{"userType", "ouId", "ouName", "ouHandle"},
		},
	}
)

type CallToRegistrationTestSuite struct {
	suite.Suite
	config       *common.TestSuiteConfig
	ouID         string
	userTypeID   string
	subFlowID    string
	authFlowID   string
	appID        string
}

func TestCallToRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(CallToRegistrationTestSuite))
}

func (ts *CallToRegistrationTestSuite) SetupSuite() {
	ts.config = &common.TestSuiteConfig{}

	ouID, err := testutils.CreateOrganizationUnit(callRegOU)
	ts.Require().NoError(err, "Failed to create OU")
	ts.ouID = ouID

	callRegUserType.OUID = ts.ouID
	utID, err := testutils.CreateUserType(callRegUserType)
	ts.Require().NoError(err, "Failed to create user type")
	ts.userTypeID = utID

	subFlowID, err := testutils.CreateFlow(callRegSubFlow)
	ts.Require().NoError(err, "Failed to create registration sub-flow")
	ts.subFlowID = subFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, subFlowID)

	// Wire the callee flow ID into the CALL node.
	authNodes, ok := callRegAuthFlow.Nodes.([]map[string]interface{})
	ts.Require().True(ok, "auth flow Nodes should be []map[string]interface{}")
	for _, node := range authNodes {
		if node["id"] == "call_reg" {
			node["flow"] = map[string]interface{}{"ref": subFlowID}
		}
	}
	authFlowID, err := testutils.CreateFlow(callRegAuthFlow)
	ts.Require().NoError(err, "Failed to create auth caller flow")
	ts.authFlowID = authFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, authFlowID)

	callRegTestApp.OUID = ts.ouID
	callRegTestApp.AuthFlowID = authFlowID
	callRegTestApp.RegistrationFlowID = subFlowID
	appID, err := testutils.CreateApplication(callRegTestApp)
	ts.Require().NoError(err, "Failed to create app")
	ts.appID = appID
}

func (ts *CallToRegistrationTestSuite) TearDownSuite() {
	if ts.appID != "" {
		if err := testutils.DeleteApplication(ts.appID); err != nil {
			ts.T().Logf("Failed to delete app: %v", err)
		}
	}
	for _, fid := range ts.config.CreatedFlowIDs {
		if err := testutils.DeleteFlow(fid); err != nil {
			ts.T().Logf("Failed to delete flow %s: %v", fid, err)
		}
	}
	if ts.userTypeID != "" {
		if err := testutils.DeleteUserType(ts.userTypeID); err != nil {
			ts.T().Logf("Failed to delete user type: %v", err)
		}
	}
	if ts.ouID != "" {
		if err := testutils.DeleteOrganizationUnit(ts.ouID); err != nil {
			ts.T().Logf("Failed to delete OU: %v", err)
		}
	}
}

// TestCallToRegistration_SuspendAndResume exercises the full CALL→callee_END→caller_resume path.
// Initiating the auth flow lands on the registration sub-flow's prompt (suspended at the
// callee). Resuming with credentials runs ProvisioningExecutor (which authenticates the new
// user), pops the frame, and runs the caller's AuthAssertExecutor to produce a JWT assertion.
func (ts *CallToRegistrationTestSuite) TestCallToRegistration_SuspendAndResume() {
	err := common.UpdateAppConfig(ts.appID, ts.authFlowID, ts.subFlowID)
	ts.Require().NoError(err, "App config update should succeed")

	// Step 1: initiate auth flow — engine pushes call frame and transitions into the
	// registration sub-flow, whose first non-START node is the prompt.
	flowStep, err := common.InitiateAuthenticationFlow(ts.appID, false, nil, "")
	ts.Require().NoError(err, "Failed to initiate auth flow")
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected INCOMPLETE at callee prompt")
	ts.Require().Equal("VIEW", flowStep.Type)
	ts.Require().NotEmpty(flowStep.ExecutionID)
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "username"))
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"))

	// Step 2: submit credentials. Provisioning creates the user and sets AuthenticatedUser
	// in the engine context (preserved across frame boundaries). On callee END, control returns
	// to the caller and auth_assert produces the assertion.
	username := common.GenerateUniqueUsername("calltoreg")
	inputs := map[string]string{
		"username": username,
		"password": "Password123!",
		"email":    username + "@example.com",
	}
	complete, err := common.CompleteFlow(flowStep.ExecutionID, inputs, "action_register", flowStep.ChallengeToken)
	ts.Require().NoError(err, "Failed to complete flow")
	ts.Require().Equal("COMPLETE", complete.FlowStatus, "Expected COMPLETE after callee return + auth_assert")
	ts.Require().NotEmpty(complete.Assertion, "JWT assertion should be returned")
	ts.Require().Nil(complete.Error)

	jwtClaims, err := testutils.ValidateJWTAssertionFields(
		complete.Assertion,
		ts.appID,
		callRegUserType.Name,
		ts.ouID,
		callRegOU.Name,
		callRegOU.Handle,
	)
	ts.Require().NoError(err, "JWT assertion fields validation failed")
	ts.Require().NotNil(jwtClaims)
}
