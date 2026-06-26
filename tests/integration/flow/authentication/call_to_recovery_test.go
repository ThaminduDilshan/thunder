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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/tests/integration/flow/common"
	"github.com/thunder-id/thunderid/tests/integration/testutils"
)

// Tests an AUTHENTICATION flow that CALLs a RECOVERY sub-flow. The sub-flow identifies
// the user and resets their credential. On callee END, the caller resumes at its END.

var (
	callRecOU = testutils.OrganizationUnit{
		Handle:      "call_to_recovery_ou",
		Name:        "Test OU for CALL→Recovery",
		Description: "Organization unit for call-to-recovery test",
	}

	callRecUserType = testutils.UserType{
		Name: "call_to_recovery_user",
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

	// Recovery sub-flow: prompt for username + new password, identify the user, reset credential.
	callRecSubFlow = testutils.Flow{
		Name:     "Call-To-Recovery Sub-Flow",
		FlowType: "RECOVERY",
		Handle:   "call_to_recovery_sub_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "prompt_recovery",
			},
			{
				"id":   "prompt_recovery",
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
								"ref":        "input_new_password",
								"identifier": "password",
								"type":       "PASSWORD_INPUT",
								"required":   true,
							},
						},
						"action": map[string]interface{}{
							"ref":      "action_recover",
							"nextNode": "identify_user",
						},
					},
				},
			},
			{
				"id":   "identify_user",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "IdentifyingExecutor",
					"mode": "identify",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_username",
							"identifier": "username",
							"type":       "TEXT_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess":    "set_credential",
				"onIncomplete": "prompt_recovery",
			},
			{
				"id":   "set_credential",
				"type": "TASK_EXECUTION",
				"executor": map[string]interface{}{
					"name": "CredentialSetter",
					"inputs": []map[string]interface{}{
						{
							"ref":        "input_new_password",
							"identifier": "password",
							"type":       "PASSWORD_INPUT",
							"required":   true,
						},
					},
				},
				"onSuccess": "end",
			},
			{
				"id":   "end",
				"type": "END",
			},
		},
	}

	// Caller auth flow: START → CALL(recovery sub-flow) → END.
	callRecAuthFlow = testutils.Flow{
		Name:     "Call-To-Recovery Auth Caller Flow",
		FlowType: "AUTHENTICATION",
		Handle:   "call_to_recovery_auth_flow",
		Nodes: []map[string]interface{}{
			{
				"id":        "start",
				"type":      "START",
				"onSuccess": "call_recovery",
			},
			{
				"id":        "call_recovery",
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

	callRecTestApp = testutils.Application{
		Name:         "Call-To-Recovery Test App",
		Description:  "App for CALL→Recovery test",
		ClientID:     "call_to_recovery_client",
		ClientSecret: "secret123",
		RedirectURIs: []string{"http://localhost:3000/callback"},
		AllowedUserTypes: []string{"call_to_recovery_user"},
		AssertionConfig: map[string]interface{}{
			"userAttributes": []string{"userType"},
		},
	}
)

type CallToRecoveryTestSuite struct {
	suite.Suite
	config     *common.TestSuiteConfig
	ouID       string
	userTypeID string
	subFlowID  string
	authFlowID string
	appID      string
	testUserID string
	testUser   string
}

func TestCallToRecoveryTestSuite(t *testing.T) {
	suite.Run(t, new(CallToRecoveryTestSuite))
}

func (ts *CallToRecoveryTestSuite) SetupSuite() {
	ts.config = &common.TestSuiteConfig{}
	ts.testUser = common.GenerateUniqueUsername("recuser")

	ouID, err := testutils.CreateOrganizationUnit(callRecOU)
	ts.Require().NoError(err, "Failed to create OU")
	ts.ouID = ouID

	callRecUserType.OUID = ts.ouID
	utID, err := testutils.CreateUserType(callRecUserType)
	ts.Require().NoError(err, "Failed to create user type")
	ts.userTypeID = utID

	userIDs, err := testutils.CreateMultipleUsers(testutils.User{
		OUID: ts.ouID,
		Type: callRecUserType.Name,
		Attributes: json.RawMessage(`{
			"username": "` + ts.testUser + `",
			"password": "InitialPwd123!"
		}`),
	})
	ts.Require().NoError(err, "Failed to create test user")
	ts.testUserID = userIDs[0]
	ts.config.CreatedUserIDs = userIDs

	subFlowID, err := testutils.CreateFlow(callRecSubFlow)
	ts.Require().NoError(err, "Failed to create recovery sub-flow")
	ts.subFlowID = subFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, subFlowID)

	// Wire the callee flow ID into the CALL node.
	authNodes, ok := callRecAuthFlow.Nodes.([]map[string]interface{})
	ts.Require().True(ok, "auth flow Nodes should be []map[string]interface{}")
	for _, node := range authNodes {
		if node["id"] == "call_recovery" {
			node["flow"] = map[string]interface{}{"ref": subFlowID}
		}
	}
	authFlowID, err := testutils.CreateFlow(callRecAuthFlow)
	ts.Require().NoError(err, "Failed to create auth caller flow")
	ts.authFlowID = authFlowID
	ts.config.CreatedFlowIDs = append(ts.config.CreatedFlowIDs, authFlowID)

	callRecTestApp.OUID = ts.ouID
	callRecTestApp.AuthFlowID = authFlowID
	appID, err := testutils.CreateApplication(callRecTestApp)
	ts.Require().NoError(err, "Failed to create app")
	ts.appID = appID
}

func (ts *CallToRecoveryTestSuite) TearDownSuite() {
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

// TestCallToRecovery_SuspendAndResume exercises an auth flow CALLing a recovery sub-flow.
// Initiate → INCOMPLETE at the recovery prompt → submit username + new password →
// IdentifyingExecutor sets userID in runtime data, CredentialSetter updates the password,
// callee reaches END, engine pops the frame, caller's onSuccess (end) → COMPLETE.
func (ts *CallToRecoveryTestSuite) TestCallToRecovery_SuspendAndResume() {
	err := common.UpdateAppConfig(ts.appID, ts.authFlowID, "")
	ts.Require().NoError(err, "App config update should succeed")

	flowStep, err := common.InitiateAuthenticationFlow(ts.appID, false, nil, "")
	ts.Require().NoError(err, "Failed to initiate auth flow")
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected INCOMPLETE at callee prompt")
	ts.Require().Equal("VIEW", flowStep.Type)
	ts.Require().NotEmpty(flowStep.ExecutionID)
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "username"))
	ts.Require().True(common.HasInput(flowStep.Data.Inputs, "password"))

	inputs := map[string]string{
		"username": ts.testUser,
		"password": "NewPwd456!",
	}
	complete, err := common.CompleteFlow(flowStep.ExecutionID, inputs, "action_recover", flowStep.ChallengeToken)
	ts.Require().NoError(err, "Failed to complete flow")
	ts.Require().Equal("COMPLETE", complete.FlowStatus, "Expected COMPLETE after callee return")
	ts.Require().Nil(complete.Error)
}
