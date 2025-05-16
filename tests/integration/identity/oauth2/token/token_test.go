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

package token

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
	clientId      = "client123"
	clientSecret  = "secret123"
)

type TokenTestSuite struct {
	suite.Suite
}

func TestTokenTestSuite(t *testing.T) {

	suite.Run(t, new(TokenTestSuite))
}

func (ts *TokenTestSuite) runClientCredentialsTestCase(request *http.Request,
	expectedStatus int, expectedScopes []string, expectedError string) {

	// Configure the HTTP client to skip TLS verification.
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Send the request.
	resp, err := client.Do(request)
	if err != nil {
		ts.T().Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Validate the response status.
	if resp.StatusCode != expectedStatus {
		ts.T().Fatalf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Parse the response body.
	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		ts.T().Fatalf("Failed to parse response body: %v", err)
	}

	// Validate the response content.
	if expectedStatus == http.StatusOK {
		if _, ok := respBody["access_token"]; !ok {
			ts.T().Fatalf("Response does not contain access_token")
		}
		if _, ok := respBody["token_type"]; !ok {
			ts.T().Fatalf("Response does not contain token_type")
		}
		if _, ok := respBody["expires_in"]; !ok {
			ts.T().Fatalf("Response does not contain expires_in")
		}
		if len(expectedScopes) > 0 {
			if _, ok := respBody["scope"]; !ok {
				ts.T().Fatalf("Response does not contain scope")
			}
			scopes := strings.Fields(respBody["scope"].(string))
			if len(scopes) != len(expectedScopes) {
				ts.T().Fatalf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
			}
			for _, expectedScope := range expectedScopes {
				found := false
				for _, scope := range scopes {
					if scope == expectedScope {
						found = true
						break
					}
				}
				if !found {
					ts.T().Fatalf("Expected scope %s not found in response", expectedScope)
				}
			}
		} else if _, ok := respBody["scope"]; ok {
			ts.T().Fatalf("Response should not contain scope when no scopes are requested")
		}
	} else if expectedStatus == http.StatusBadRequest {
		if _, ok := respBody["error"]; !ok {
			ts.T().Fatalf("Response does not contain error")
		}
		if respBody["error"] != expectedError {
			ts.T().Fatalf("Expected error '%s', got '%v'", expectedError, respBody["error"])
		}
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantWithHeaderCredentials() {

	testCases := []struct {
		testName        string
		requestedScopes string
		expectedStatus  int
		expectedScopes  []string
	}{
		{
			testName:        "WithAuthorizedScopes",
			requestedScopes: "internal_user_mgt_view internal_user_mgt_edit internal_group_mgt_view",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "internal_user_mgt_edit", "internal_group_mgt_view"},
		},
		{
			testName:        "WithoutScopes",
			requestedScopes: "",
			expectedStatus:  http.StatusOK,
			expectedScopes:  nil,
		},
		{
			testName:        "WithAuthorizedAndUnauthorizedScopes",
			requestedScopes: "internal_user_mgt_view internal_group_mgt_view internal_group_mgt_edit",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "internal_group_mgt_view"},
		},
		{
			testName:        "WithInvalidScopes",
			requestedScopes: "invalid_scope",
			expectedStatus:  http.StatusOK,
			expectedScopes:  nil,
		},
		{
			testName:        "WithAuthorizedAndInvalidScopes",
			requestedScopes: "internal_user_mgt_view invalid_scope",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view"},
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			// Prepare the request.
			reqBody := strings.NewReader("grant_type=client_credentials&scope=" + tc.requestedScopes)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.SetBasicAuth(clientId, clientSecret)

			// Run the test.
			ts.runClientCredentialsTestCase(request, tc.expectedStatus, tc.expectedScopes, "")
		})
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantWithBodyCredentials() {

	testCases := []struct {
		testName        string
		requestedScopes string
		expectedStatus  int
		expectedScopes  []string
	}{
		{
			testName:        "WithAuthorizedScopes",
			requestedScopes: "internal_user_mgt_view internal_user_mgt_edit",
			expectedStatus:  http.StatusOK,
			expectedScopes:  []string{"internal_user_mgt_view", "internal_user_mgt_edit"},
		},
		{
			testName:        "WithoutScopes",
			requestedScopes: "",
			expectedStatus:  http.StatusOK,
			expectedScopes:  nil,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			reqBody := strings.NewReader("grant_type=client_credentials&scope=" + tc.requestedScopes +
				"&client_id=" + clientId + "&client_secret=" + clientSecret)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			ts.runClientCredentialsTestCase(request, tc.expectedStatus, tc.expectedScopes, "")
		})
	}
}

func (ts *TokenTestSuite) TestClientCredentialsGrantNegativeCases() {

	testCases := []struct {
		testName       string
		requestBody    string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			testName:       "InvalidHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic " + basicAuth("invalid", "invalid"),
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "IncorrectHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic invalid_base64",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "InvalidHeaderCredentials",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "Basic invalid_base64",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "InvalidCredentialsInBody",
			requestBody:    "grant_type=client_credentials&client_id=invalid&client_secret=invalid",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid_client",
		},
		{
			testName:       "MissingCredentialsInBody",
			requestBody:    "grant_type=client_credentials",
			authHeader:     "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
		{
			testName:       "InvalidGrantType",
			requestBody:    "grant_type=invalid_grant",
			authHeader:     "Basic " + basicAuth(clientId, clientSecret),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "unsupported_grant_type",
		},
		{
			testName:       "MissingGrantType",
			requestBody:    "",
			authHeader:     "Basic " + basicAuth(clientId, clientSecret),
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid_request",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			// Prepare the request.
			reqBody := strings.NewReader(tc.requestBody)
			request, err := http.NewRequest("POST", testServerURL+"/oauth2/token", reqBody)
			if err != nil {
				ts.T().Fatalf("Failed to create request: %v", err)
			}
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			if tc.authHeader != "" {
				request.Header.Set("Authorization", tc.authHeader)
			}

			// Run the test.
			ts.runClientCredentialsTestCase(request, tc.expectedStatus, nil, tc.expectedError)
		})
	}
}

func (ts *TokenTestSuite) TestAuthorizationCodeGrantFlow() {

	testCases := []struct {
		testName       string
		username       string
		password       string
		expectedStatus int
		expectedError  string
	}{
		{
			testName:       "ValidAuthorizationCodeFlow",
			username:       "thor",
			password:       "thor123",
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			testName:       "InvalidCredentials",
			username:       "thor",
			password:       "invalid",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "access_denied",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.testName, func() {
			// Step 1: Initiate the authorization request.
			requestURI := testServerURL + "/oauth2/authorize?response_type=code&client_id=" + clientId +
				"&redirect_uri=https://localhost:3000&scope=openid&state=state_1"
			authRequest, err := http.NewRequest("GET", requestURI, nil)
			if err != nil {
				ts.T().Fatalf("Failed to create authorization request: %v", err)
			}

			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
			}

			authResponse, err := client.Do(authRequest)
			if err != nil {
				ts.T().Fatalf("Failed to send authorization request: %v", err)
			}
			defer authResponse.Body.Close()

			// TODO: This is a redirect and cannot extract sessionDataKey from test suite.
			//  Need to find a way. Need to check on whether it's permitted to send a location header.

			// Validate the authorization response.
			if authResponse.StatusCode != http.StatusOK {
				ts.T().Fatalf("Expected status %d, got %d", http.StatusOK, authResponse.StatusCode)
			}

			// Extract sessionDataKey from the Location header.
			redirectURL := authResponse.Header.Get("Location")
			if !strings.Contains(redirectURL, "sessionDataKey=") {
				ts.T().Fatalf("Redirect URL does not contain sessionDataKey")
			}
			sessionDataKey := strings.Split(strings.Split(redirectURL, "sessionDataKey=")[1], "&")[0]

			// Step 2: Authenticate the user.
			authnBody := strings.NewReader("username=" + tc.username + "&password=" + tc.password +
				"&sessionDataKey=" + sessionDataKey)
			authnRequest, err := http.NewRequest("POST", testServerURL+"/flow/authn", authnBody)
			if err != nil {
				ts.T().Fatalf("Failed to create authentication request: %v", err)
			}
			authnRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			authnResponse, err := client.Do(authnRequest)
			if err != nil {
				ts.T().Fatalf("Failed to send authentication request: %v", err)
			}
			defer authnResponse.Body.Close()

			// Validate the authentication response.
			if authnResponse.StatusCode != http.StatusFound {
				ts.T().Fatalf("Expected status %d, got %d", http.StatusFound, authnResponse.StatusCode)
			}

			// Extract sessionDataKey from the redirect response.
			redirectURL = authnResponse.Header.Get("Location")
			if !strings.Contains(redirectURL, "sessionDataKey=") {
				ts.T().Fatalf("Redirect URL does not contain sessionDataKey")
			}
			sessionDataKey = strings.Split(strings.Split(redirectURL, "sessionDataKey=")[1], "&")[0]

			// Step 3: Send another authorize request with the new sessionDataKey.
			requestURI = testServerURL + "/oauth2/authorize?sessionDataKey=" + sessionDataKey
			authRequest, err = http.NewRequest("GET", requestURI, nil)
			if err != nil {
				ts.T().Fatalf("Failed to create authorization request: %v", err)
			}

			authResponse, err = client.Do(authRequest)
			if err != nil {
				ts.T().Fatalf("Failed to send authorization request: %v", err)
			}
			defer authResponse.Body.Close()

			// Validate the final redirect response.
			if authResponse.StatusCode != http.StatusFound {
				ts.T().Fatalf("Expected status %d, got %d", http.StatusFound, authResponse.StatusCode)
			}

			redirectURL = authResponse.Header.Get("Location")
			if !strings.Contains(redirectURL, "code=") {
				ts.T().Fatalf("Redirect URL does not contain authorization code")
			}
			code := strings.Split(strings.Split(redirectURL, "code=")[1], "&")[0]

			// Step 4: Exchange the authorization code for an access token.
			tokenBody := strings.NewReader("grant_type=authorization_code&code=" + code +
				"&redirect_uri=https://localhost:3000&client_id=" + clientId + "&client_secret=" + clientSecret)
			tokenRequest, err := http.NewRequest("POST", testServerURL+"/oauth2/token", tokenBody)
			if err != nil {
				ts.T().Fatalf("Failed to create token request: %v", err)
			}
			tokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			tokenResponse, err := client.Do(tokenRequest)
			if err != nil {
				ts.T().Fatalf("Failed to send token request: %v", err)
			}
			defer tokenResponse.Body.Close()

			// Validate the token response.
			if tokenResponse.StatusCode != http.StatusOK {
				ts.T().Fatalf("Expected status %d, got %d", http.StatusOK, tokenResponse.StatusCode)
			}

			var tokenRespBody map[string]interface{}
			err = json.NewDecoder(tokenResponse.Body).Decode(&tokenRespBody)
			if err != nil {
				ts.T().Fatalf("Failed to parse token response body: %v", err)
			}

			if _, ok := tokenRespBody["access_token"]; !ok {
				ts.T().Fatalf("Response does not contain access_token")
			}
			if _, ok := tokenRespBody["token_type"]; !ok {
				ts.T().Fatalf("Response does not contain token_type")
			}
			if _, ok := tokenRespBody["expires_in"]; !ok {
				ts.T().Fatalf("Response does not contain expires_in")
			}
		})
	}
}

func basicAuth(username, password string) string {

	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
