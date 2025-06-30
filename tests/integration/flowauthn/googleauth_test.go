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

package flowauthn

import (
	"net/url"
	"strings"
	"testing"

	"github.com/asgardeo/thunder/tests/integration/testutils"
	"github.com/stretchr/testify/suite"
)

type GoogleAuthFlowTestSuite struct {
	suite.Suite
	mockGoogleIDP *testutils.MockGoogleIDP
	testAppID     string
	testIdpID     string
}

func TestGoogleAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(GoogleAuthFlowTestSuite))
}

func (ts *GoogleAuthFlowTestSuite) SetupSuite() {
	// Set up mock Google IDP server
	var err error
	ts.mockGoogleIDP, err = testutils.NewMockGoogleIDP()
	if err != nil {
		ts.T().Fatalf("Failed to create mock Google IDP: %v", err)
	}

	// Create test Google IDP with mock endpoints
	ts.testIdpID, err = createTestGoogleIDP(ts.mockGoogleIDP)
	if err != nil {
		ts.T().Fatalf("Failed to create test Google IDP: %v", err)
	}
	if ts.testIdpID == "" {
		ts.T().Fatal("Test Google IDP creation returned empty ID")
	}
	ts.T().Logf("Test Google IDP created successfully")

	// Create test application for Google auth
	ts.testAppID, err = createTestGoogleApp()
	if err != nil {
		ts.T().Fatalf("Failed to create test Google application: %v", err)
	}
	if ts.testAppID == "" {
		ts.T().Fatal("Test Google application creation returned empty ID")
	}
	ts.T().Logf("Test Google application created successfully")
}

func (ts *GoogleAuthFlowTestSuite) TearDownSuite() {
	ts.T().Logf("Cleaning up Google Auth Flow Test resources")

	// Clean up mock IDP server
	if ts.mockGoogleIDP != nil {
		ts.mockGoogleIDP.Close()
		ts.T().Logf("Mock Google IDP server stopped")
	}

	// Delete the test application
	err := deleteTestApp(ts.testAppID)
	if err != nil {
		ts.T().Logf("Warning: Failed to delete test Google app: %v", err)
	} else {
		ts.T().Logf("Test Google application deleted")
	}

	// Delete the test Google IDP
	err = deleteTestGoogleIDP(ts.testIdpID)
	if err != nil {
		ts.T().Logf("Warning: Failed to delete test Google IDP: %v", err)
	} else {
		ts.T().Logf("Test Google IDP deleted")
	}
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInitiation() {
	// Initialize the flow by calling the flow execution API
	flowStep, err := initiateAuthFlow(ts.testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	// Verify flow status and type
	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Validate redirect information
	ts.Require().NotEmpty(flowStep.Data, "Flow data should not be empty")
	ts.Require().NotEmpty(flowStep.Data.RedirectURL, "Redirect URL should not be empty")
	redirectURLStr := flowStep.Data.RedirectURL
	ts.Require().True(strings.HasPrefix(redirectURLStr, "https://accounts.google.com/o/oauth2/v2/auth"),
		"Redirect URL should point to Google authentication")

	// // Should redirect to mock Google IDP instead of real Google
	// expectedPrefix := ts.mockGoogleIDP.GetAuthorizationURL()
	// ts.Require().True(strings.HasPrefix(redirectURLStr, expectedPrefix),
	// 	"Redirect URL should point to mock Google authentication server. Expected prefix: %s, Got: %s",
	// 	expectedPrefix, redirectURLStr)

	// Parse and validate the redirect URL
	redirectURL, err := url.Parse(redirectURLStr)
	ts.Require().NoError(err, "Should be able to parse the redirect URL")

	// Check required query parameters in the redirect URL
	queryParams := redirectURL.Query()
	ts.Require().NotEmpty(queryParams.Get("client_id"), "client_id should be present in redirect URL")
	ts.Require().NotEmpty(queryParams.Get("redirect_uri"), "redirect_uri should be present in redirect URL")
	ts.Require().NotEmpty(queryParams.Get("response_type"), "response_type should be present in redirect URL")
	ts.Require().Equal("code", queryParams.Get("response_type"), "response_type should be 'code'")

	scope := queryParams.Get("scope")
	ts.Require().NotEmpty(scope, "scope should be present in redirect URL")

	scopesPresent := strings.Contains(scope, "openid") &&
		strings.Contains(scope, "email") &&
		strings.Contains(scope, "profile")
	ts.Require().True(scopesPresent, "scope should include expected scopes")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInvalidAppID() {
	errorResp, err := initiateAuthFlowWithError("invalid-google-app-id", nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate authentication flow with invalid app ID: %v", err)
	}

	ts.Require().Equal("FES-60003", errorResp.Code, "Expected error code for invalid app ID")
	ts.Require().Equal("Invalid request", errorResp.Message, "Expected error message for invalid request")
	ts.Require().Equal("Invalid app ID provided in the request", errorResp.Description,
		"Expected error description for invalid app ID")
}

func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowComplete() {
	// Step 1: Initiate the flow
	flowStep, err := initiateAuthFlow(ts.testAppID, nil)
	if err != nil {
		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
	}

	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus, "Expected flow status to be INCOMPLETE")
	ts.Require().Equal("REDIRECTION", flowStep.Type, "Expected flow type to be REDIRECT")
	ts.Require().NotEmpty(flowStep.FlowID, "Flow ID should not be empty")

	// Step 2: Simulate user completing OAuth flow and returning with authorization code
	inputs := map[string]string{
		"code": "test_auth_code_12345",
	}

	completedFlowStep, err := completeAuthFlow(flowStep.FlowID, inputs)
	if err != nil {
		ts.T().Fatalf("Failed to complete Google authentication flow: %v", err)
	}

	// Verify the flow completed successfully
	ts.Require().Equal("COMPLETE", completedFlowStep.FlowStatus, "Expected flow status to be COMPLETE")
	ts.Require().NotEmpty(completedFlowStep.Assertion, "Expected assertion to be present")

	// The assertion should be a JWT token - basic validation
	parts := strings.Split(completedFlowStep.Assertion, ".")
	ts.Require().Equal(3, len(parts), "Assertion should be a valid JWT with 3 parts")
}

// func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowInvalidCode() {
// 	// Step 1: Initiate the flow
// 	flowStep, err := initiateAuthFlow(testGoogleAppID, nil)
// 	if err != nil {
// 		ts.T().Fatalf("Failed to initiate Google authentication flow: %v", err)
// 	}

// 	// Step 2: Try to complete flow with invalid authorization code
// 	inputs := map[string]string{
// 		"code": "invalid_authorization_code",
// 	}

// 	errorResp, err := completeAuthFlowWithError(flowStep.FlowID, inputs)
// 	if err != nil {
// 		ts.T().Fatalf("Failed to complete authentication flow with invalid code: %v", err)
// 	}

// 	// Should get an error for invalid authorization code
// 	ts.Require().NotEmpty(errorResp.Code, "Expected error code for invalid authorization code")
// 	ts.Require().NotEmpty(errorResp.Message, "Expected error message for invalid authorization code")
// }

// func (ts *GoogleAuthFlowTestSuite) TestGoogleAuthFlowResourceIsolation() {
// 	// Verify that our test created its own resources and doesn't use system defaults

// 	// 1. Verify test app exists and is different from default system app
// 	testAppConfig, err := getAppConfig(testGoogleAppID)
// 	if err != nil {
// 		ts.T().Fatalf("Failed to get test app config: %v", err)
// 	}

// 	ts.Require().Equal("test-google-app-12345", testAppConfig["id"])
// 	ts.Require().Equal("Google Auth Test App", testAppConfig["name"])
// 	ts.Require().Equal("auth_flow_config_google", testAppConfig["auth_flow_graph_id"])

// 	// 2. Verify system default app is unaffected
// 	systemAppConfig, err := getAppConfig(appID) // This is the default system app
// 	if err != nil {
// 		ts.T().Fatalf("Failed to get system app config: %v", err)
// 	}

// 	// System app should still have basic auth flow (not modified by our test)
// 	ts.Require().NotEqual("auth_flow_config_google", systemAppConfig["auth_flow_graph_id"],
// 		"System app should not be modified by Google auth test")

// 	// 3. Test that we can initiate flow with our test app
// 	flowStep, err := initiateAuthFlow(testGoogleAppID, nil)
// 	if err != nil {
// 		ts.T().Fatalf("Failed to initiate flow with test app: %v", err)
// 	}

// 	ts.Require().Equal("INCOMPLETE", flowStep.FlowStatus)
// 	ts.Require().Equal("REDIRECTION", flowStep.Type)

// 	// Should redirect to our mock IDP
// 	expectedPrefix := ts.mockGoogleIDP.GetAuthorizationURL()
// 	ts.Require().True(strings.HasPrefix(flowStep.Data.RedirectURL, expectedPrefix),
// 		"Should redirect to mock IDP, not real Google")
// }
