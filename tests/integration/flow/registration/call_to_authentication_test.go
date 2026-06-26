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

package registration

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/flow/common"
	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

// Tests a REGISTRATION flow that CALLs an AUTHENTICATION sub-flow. Useful for journeys where
// a user provides an identifier already known to the system — the registration flow then
// delegates to an authentication sub-flow to log the existing user in instead of provisioning.

var (
	callAuthOU = testutils.OrganizationUnit{
		Handle:      "call_to_auth_ou",
		Name:        "Test OU for CALL→Authentication",
		Description: "Organization unit for call-to-authentication test",
	}

	callAuthUserType = testutils.UserType{
		Name: "call_to_auth_user",
		Schema: map[string]interface{}{
			"username": map[string]interface{}{
				"type": "string",
			},
			"password": map[string]interface{}{
				"type":       "string",
				"credential": true,
			},
		},
	}

	// Authentication sub-flow: prompts for username + password and runs BasicAuthExecutor.
	callAuthSubFlow = testutils.Flow{
		Name:     "Call-To-Auth Auth Sub-Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "call_to_auth_sub_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_creds",
			},
			{
				"id":   "prompt_creds",
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
						},
						"action": map[string]interface{}{
							"ref":      "action_signin",
							"nextNode": "basic_auth",
						},
					},
				},
			},
			{
				"id":   "basic_auth",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "BasicAuthExecutor",
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
					},
				},
				"onSuccess":    "end",
				"onIncomplete": "prompt_creds",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	// Caller registration flow: START → CALL(auth sub-flow) → END.
	// Demonstrates a registration journey that delegates to an authentication flow.
	callAuthRegFlow = testutils.Flow{
		Name:     "Call-To-Auth Reg Caller Flow",
		FlowType: "REGISTRATION",
		Handle:   "call_to_auth_reg_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "call_auth",
			},
			{
				"id":        "call_auth",
				"type":      "CALL",
				"flow":      map[string]interface{}{"ref": ""}, // populated in SetupSuite
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	callAuthTestApp = testutils.Application{
		Name:                      "Call-To-Auth Test App",
		Description:               "App for CALL→Authentication test",
		IsRegistrationFlowEnabled: true,
		ClientID:                  "call_to_auth_client",
		ClientSecret:              "secret123",
		RedirectURIs:              []string{"http://localhost:3000/callback"},
		AllowedUserTypes:          []string{"call_to_auth_user"},
		AssertionConfig: map[string]interface{}{
			"userAttributes": []string{"userType"},
		},
	}
)

type CallToAuthenticationTestSuite struct {
	suite.Suite
	config     *common.TestSuiteConfig
	ouID       string
	userTypeID string
	subFlowID  string
	regFlowID  string
	appID      string
	testUser   string
	testPass   string
}

func TestCallToAuthenticationTestSuite(t *testing.T) {
	suite.Run(t, new(CallToAuthenticationTestSuite))
}

func (ts *CallToAuthenticationTestSuite) SetupSuite() {
	ts.config = &common.TestSuiteConfig{}
	ts.testUser = common.GenerateUniqueUsername("callauth")
	ts.testPass = "ExistingPwd123!"

	ouID, err := testutils.CreateOrganizationUnit(callAuthOU)
	ts.Require().NoError(err, "Failed to create OU")
	ts.ouID = ouID

	callAuthUserType.OUID = ts.ouID
	utID, err := testutils.CreateUserType(callAuthUserType)
	ts.Require().NoError(err, "Failed to create user type")
	ts.userTypeID = utID

	userIDs, err := testutils.CreateMultipleUsers(testutils.User{
		OUID: ts.ouID,
		Type: callAuthUserType.Name,
		Attributes: json.RawMessage(`{
			"username": "` + ts.testUser + `",
			"password": "` + ts.testPass + `"
		}`),
	})
	ts.Require().NoError(err, "Failed to create test user")
	ts.config.CreatedUserIDs = userIDs

	subFlowID, err := testutils.CreateFlow(callAuthSubFlow)
	ts.Require().NoError(err, "Failed to create auth sub-flow")
	ts.subFlowID = subFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, subFlowID)

	// Wire the callee flow ID into the CALL node.
	regNodes, ok := callAuthRegFlow.Nodes.([]map[string]interface{})
	ts.Require().True(ok, "reg flow Nodes should be []map[string]interface{}")
	for _, node := range regNodes {
		if node["id"] == "call_auth" {
			node["flow"] = map[string]interface{}{"ref": subFlowID}
		}
	}
	regFlowID, err := testutils.CreateFlow(callAuthRegFlow)
	ts.Require().NoError(err, "Failed to create registration caller flow")
	ts.regFlowID = regFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, regFlowID)

	callAuthTestApp.OUID = ts.ouID
	callAuthTestApp.RegistrationFlowID = regFlowID
	appID, err := testutils.CreateApplication(callAuthTestApp)
	ts.Require().NoError(err, "Failed to create app")
	ts.appID = appID
}

func (ts *CallToAuthenticationTestSuite) TearDownSuite() {
	if err := testutils.CleanupUsers(ts.config.CreatedUserIDs); err != nil {
		ts.T().Logf("Failed to cleanup users: %v", err)
	}
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

// TestCallToAuthentication_SuspendAndResume exercises a registration flow CALLing an
// authentication sub-flow. Initiate → INCOMPLETE at the callee credentials prompt →
// submit credentials → BasicAuthExecutor authenticates → callee END → engine pops the
// frame → caller's onSuccess (end) → COMPLETE.
func (ts *CallToAuthenticationTestSuite) TestCallToAuthentication_SuspendAndResume() {
	err := common.UpdateAppConfig(ts.appID, "", ts.regFlowID)
	ts.Require().NoError(err, "App config update should succeed")

	flowStep, err := common.InitiateRegistrationFlow(ts.appID, false, nil, "")
	ts.Require().NoError(err, "Failed to initiate registration flow")
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected INCOMPLETE at callee prompt")
	ts.Require().Equal("VIEW", flowStep.Type)
	ts.Require().NotEmpty(flowStep.ExecutionID)
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "username"))
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"))

	inputs := map[string]string{
		"username": ts.testUser,
		"password": ts.testPass,
	}
	complete, err := common.CompleteFlow(flowStep.ExecutionID, inputs, "action_signin", flowStep.ChallengeToken)
	ts.Require().NoError(err, "Failed to complete flow")
	ts.Require().Equal("COMPLETE", complete.FlowStatus, "Expected COMPLETE after callee return")
	ts.Require().Nil(complete.Error)
}
