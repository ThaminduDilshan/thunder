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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/asgardeo/thunder/tests/integration/testutils"
)

const testServerURL = "https://localhost:8095"

// Helper function to initiate the authentication flow
func initiateAuthFlow(appID string, inputs map[string]string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
	}
	if len(inputs) > 0 {
		flowReqBody["inputs"] = inputs
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var flowStep FlowStep
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &flowStep, nil
}

// Helper function to initiate the authentication flow with error handling
func initiateAuthFlowWithError(appID string, inputs map[string]string) (*ErrorResponse, error) {
	flowReqBody := map[string]interface{}{
		"applicationId": appID,
	}
	if len(inputs) > 0 {
		flowReqBody["inputs"] = inputs
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var errorResponse ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse error response body: %w", err)
	}

	return &errorResponse, nil
}

// Helper function to complete the authentication flow
func completeAuthFlow(flowID string, inputs map[string]string) (*FlowStep, error) {
	flowReqBody := map[string]interface{}{
		"flowId": flowID,
		"inputs": inputs,
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var flowStep FlowStep
	err = json.NewDecoder(resp.Body).Decode(&flowStep)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}

	return &flowStep, nil
}

// Helper function to complete the authentication flow with error handling
func completeAuthFlowWithError(flowID string, inputs map[string]string) (*ErrorResponse, error) {
	flowReqBody := map[string]interface{}{
		"flowId": flowID,
		"inputs": inputs,
	}

	reqBody, err := json.Marshal(flowReqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", testServerURL+"/flow/execute", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create flow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send flow request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var errorResponse ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse error response body: %w", err)
	}

	return &errorResponse, nil
}

// Helper function to create a user
func createUser(user User) (string, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	reqBody := bytes.NewReader(userJSON)
	req, err := http.NewRequest("POST", testServerURL+"/users", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var respBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse response body: %w", err)
	}

	id, ok := respBody["id"].(string)
	if !ok {
		return "", fmt.Errorf("response does not contain id")
	}
	return id, nil
}

// Helper function to delete a user
func deleteUser(userID string) error {
	req, err := http.NewRequest("DELETE", testServerURL+"/users/"+userID, nil)
	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete user, status code: %d", resp.StatusCode)
	}
	return nil
}

// getAppConfig retrieves the current application configuration
func getAppConfig(appID string) (map[string]interface{}, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var appConfig map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&appConfig); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	return appConfig, nil
}

// updateAppConfigWithFlow updates the application configuration with the specified auth flow graph ID
func updateAppConfigWithFlow(appID string, authFlowGraphID string) error {
	appConfig, err := getAppConfig(appID)
	if err != nil {
		return fmt.Errorf("failed to get current app config: %w", err)
	}

	appConfig["auth_flow_graph_id"] = authFlowGraphID
	appConfig["client_secret"] = "secret123"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(appConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateAppConfig updates the application configuration with the provided config object
func updateAppConfig(appID string, appConfig map[string]interface{}) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(appConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON payload: %w", err)
	}

	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IdentityProvider represents an identity provider
type IdentityProvider struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Properties  []IDPProperty `json:"properties"`
}

// IDPProperty represents a property of an identity provider
type IDPProperty struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// createOrUpdateGoogleIDP creates or updates a Google IDP with mock endpoints
func createOrUpdateGoogleIDP(idpID string, mockIDP *testutils.MockGoogleIDP) error {
	googleIDP := IdentityProvider{
		ID:          idpID,
		Name:        "Mock Google IDP",
		Description: "Mock Google Identity Provider for testing",
		Properties: []IDPProperty{
			{Name: "client_id", Value: mockIDP.GetClientID(), IsSecret: false},
			{Name: "client_secret", Value: mockIDP.ClientSecret, IsSecret: true},
			{Name: "scopes", Value: "openid email profile", IsSecret: false},
		},
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(googleIDP)
	if err != nil {
		return fmt.Errorf("failed to marshal IDP JSON payload: %w", err)
	}

	// Try to update existing IDP first
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf("%s/identity-providers/%s", testServerURL, idpID),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return fmt.Errorf("failed to create IDP update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("IDP update request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// If update failed with 404, try to create the IDP
	if resp.StatusCode == http.StatusNotFound {
		req, err = http.NewRequest(
			"POST",
			fmt.Sprintf("%s/identity-providers", testServerURL),
			bytes.NewBuffer(jsonPayload),
		)
		if err != nil {
			return fmt.Errorf("failed to create IDP creation request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("IDP creation request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("IDP creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		return nil
	}

	// Update failed with error other than 404
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("IDP update failed with status %d: %s", resp.StatusCode, string(body))
}

// deleteGoogleIDP deletes the Google IDP
func deleteGoogleIDP(idpID string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/identity-providers/%s", testServerURL, idpID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create IDP delete request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("IDP delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IDP delete failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// updateAppForGoogleAuth updates the application to use Google authentication flow
func updateAppForGoogleAuth(appID string, authFlowGraphID string) error {
	appConfig, err := getAppConfig(appID)
	if err != nil {
		return fmt.Errorf("failed to get current app config: %w", err)
	}

	appConfig["auth_flow_graph_id"] = authFlowGraphID
	appConfig["client_secret"] = "secret123"

	return updateAppConfig(appID, appConfig)
}

// Application represents an application
type Application struct {
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	AuthFlowGraphID     string   `json:"auth_flow_graph_id"`
	ClientID            string   `json:"client_id"`
	ClientSecret        string   `json:"client_secret"`
	CallbackURL         []string `json:"callback_url"`
	SupportedGrantTypes []string `json:"supported_grant_types"`
}

// createTestGoogleApp creates a new application for Google auth testing
func createTestGoogleApp() (string, error) {
	testApp := Application{
		Name:            "Google Auth Test App",
		Description:     "Test application for Google authentication flow testing",
		ClientID:        "test-google-app-client-id",
		ClientSecret:    "test-google-app-client-secret",
		CallbackURL:     []string{"https://localhost:3000"},
		AuthFlowGraphID: "auth_flow_config_google",
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(testApp)
	if err != nil {
		return "", fmt.Errorf("failed to marshal app JSON payload: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/applications", testServerURL),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create app creation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("app creation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("app creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read id of the created app from the response
	var appResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&appResponse); err != nil {
		return "", fmt.Errorf("failed to parse app creation response: %w", err)
	}
	if appResponse.ID == "" {
		return "", fmt.Errorf("app creation response does not contain ID")
	}

	return appResponse.ID, nil
}

// deleteTestApp deletes a test application
func deleteTestApp(appID string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/applications/%s", testServerURL, appID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create app deletion request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("app deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("app deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// createTestGoogleIDP creates a new Google IDP for testing
func createTestGoogleIDP(mockIDP *testutils.MockGoogleIDP) (string, error) {
	googleIDP := IdentityProvider{
		Name:        "Test Google IDP",
		Description: "Test Google Identity Provider for Google auth flow testing",
		Properties: []IDPProperty{
			{Name: "client_id", Value: mockIDP.GetClientID(), IsSecret: false},
			{Name: "client_secret", Value: mockIDP.ClientSecret, IsSecret: true},
			{Name: "redirect_uri", Value: "https://localhost:3000", IsSecret: false},
			{Name: "scopes", Value: "openid email profile", IsSecret: false},
		},
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	jsonPayload, err := json.Marshal(googleIDP)
	if err != nil {
		return "", fmt.Errorf("failed to marshal IDP JSON payload: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/identity-providers", testServerURL),
		bytes.NewBuffer(jsonPayload),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create IDP creation request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("IDP creation request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("IDP creation failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read id of the created IDP from the response
	var idpResponse struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&idpResponse); err != nil {
		return "", fmt.Errorf("failed to parse IDP creation response: %w", err)
	}

	return idpResponse.ID, nil
}

// deleteTestGoogleIDP deletes a test Google IDP
func deleteTestGoogleIDP(idpID string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("%s/identity-providers/%s", testServerURL, idpID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create IDP deletion request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("IDP deletion request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IDP deletion failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
