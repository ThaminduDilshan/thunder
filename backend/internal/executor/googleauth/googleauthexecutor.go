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

// Package googleauth provides the Google OIDC authentication executor.
package googleauth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	authnmodel "github.com/asgardeo/thunder/internal/authn/model"
	"github.com/asgardeo/thunder/internal/executor/oidcauth"
	"github.com/asgardeo/thunder/internal/executor/oidcauth/model"
	flowconst "github.com/asgardeo/thunder/internal/flow/constants"
	flowmodel "github.com/asgardeo/thunder/internal/flow/model"
	"github.com/asgardeo/thunder/internal/system/constants"
	"github.com/asgardeo/thunder/internal/system/log"
	systemutils "github.com/asgardeo/thunder/internal/system/utils"
)

const loggerComponentName = "GoogleAuthExecutor"

// GoogleOIDCAuthExecutor implements the OIDC authentication executor for Google.
type GoogleOIDCAuthExecutor struct {
	*oidcauth.OIDCAuthExecutor
}

// NewGoogleOIDCAuthExecutorFromProps creates a new instance of GoogleOIDCAuthExecutor with the provided properties.
func NewGoogleOIDCAuthExecutorFromProps(execProps flowmodel.ExecutorProperties,
	oidcProps *model.BasicOIDCExecProperties) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the complete OIDC properties for Google
	compOIDCProps := &model.OIDCExecProperties{
		AuthorizationEndpoint: googleAuthorizeEndpoint,
		TokenEndpoint:         googleTokenEndpoint,
		UserInfoEndpoint:      googleUserInfoEndpoint,
		ClientID:              oidcProps.ClientID,
		ClientSecret:          oidcProps.ClientSecret,
		RedirectURI:           oidcProps.RedirectURI,
		Scopes:                oidcProps.Scopes,
		AdditionalParams:      oidcProps.AdditionalParams,
	}

	base := oidcauth.NewOIDCAuthExecutor("google_oidc_auth_executor", execProps.Name, compOIDCProps)

	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor: exec,
	}
}

// NewGoogleOIDCAuthExecutor creates a new instance of GoogleOIDCAuthExecutor with the provided details.
func NewGoogleOIDCAuthExecutor(id, name, clientID, clientSecret, redirectURI string,
	scopes []string, additionalParams map[string]string) oidcauth.OIDCAuthExecutorInterface {
	// Prepare the OIDC properties for Google
	oidcProps := &model.OIDCExecProperties{
		AuthorizationEndpoint: googleAuthorizeEndpoint,
		TokenEndpoint:         googleTokenEndpoint,
		UserInfoEndpoint:      googleUserInfoEndpoint,
		ClientID:              clientID,
		ClientSecret:          clientSecret,
		RedirectURI:           redirectURI,
		Scopes:                scopes,
		AdditionalParams:      additionalParams,
	}

	base := oidcauth.NewOIDCAuthExecutor(id, name, oidcProps)

	exec, ok := base.(*oidcauth.OIDCAuthExecutor)
	if !ok {
		panic("failed to cast GoogleOIDCAuthExecutor to OIDCAuthExecutor")
	}
	return &GoogleOIDCAuthExecutor{
		OIDCAuthExecutor: exec,
	}
}

// Execute executes the Google OIDC authentication flow.
func (g *GoogleOIDCAuthExecutor) Execute(ctx *flowmodel.FlowContext) (*flowmodel.ExecutorResponse, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Executing Google OIDC auth executor",
		log.String("executorID", g.GetID()), log.String("flowID", ctx.FlowID))

	execResp := &flowmodel.ExecutorResponse{
		Status: flowconst.ExecIncomplete,
	}

	// Check if the required input data is provided
	if g.requiredInputData(ctx, execResp) {
		// If required input data is not provided, return incomplete status with redirection to Google.
		logger.Debug("Required input data for Google OIDC auth executor is not provided")

		g.BuildAuthorizeFlow(ctx, execResp)

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)))
	} else {
		g.ProcessAuthFlowResponse(ctx, execResp)

		logger.Debug("Google OIDC auth executor execution completed",
			log.String("status", string(execResp.Status)),
			log.Bool("isAuthenticated", ctx.AuthenticatedUser.IsAuthenticated))
	}

	return execResp, nil
}

// ProcessAuthFlowResponse processes the response from the Google OIDC authentication flow.
func (g *GoogleOIDCAuthExecutor) ProcessAuthFlowResponse(ctx *flowmodel.FlowContext,
	execResp *flowmodel.ExecutorResponse) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName),
		log.String("executorID", g.GetID()), log.String("flowID", ctx.FlowID))
	logger.Debug("Processing Google OIDC auth flow response")

	execResp.Status = flowconst.ExecIncomplete

	// Process authorization code if available
	code, ok := ctx.UserInputData["code"]
	if ok && code != "" {
		// Exchange authorization code for tokenResp
		tokenResp, err := g.ExchangeCodeForToken(ctx, code)
		if err != nil {
			logger.Error("Failed to exchange code for a token", log.Error(err))
			execResp.Status = flowconst.ExecError
			execResp.Error = "Failed to authenticate with OIDC provider: " + err.Error()
			return
		}

		// Validate the token response
		if tokenResp.AccessToken == "" {
			logger.Debug("Access token is empty in the token response")
			execResp.Status = flowconst.ExecUserError
			execResp.Error = "Access token is empty in the token response. Please provide a valid authorization code."
			return
		}

		// Get user info using the access token
		userInfo, err := g.GetUserInfo(ctx, tokenResp.AccessToken)
		if err != nil {
			logger.Error("Failed to get user info", log.Error(err))
			execResp.Status = flowconst.ExecUserError
			execResp.Error = "Failed to get user information: " + err.Error()
			return
		}

		// Populate authenticated user from user info
		email := userInfo["email"]
		name := userInfo["name"]
		sub := userInfo["sub"]

		// Determine the username (prefer email over name over sub)
		username := sub
		if email != "" {
			username = email
		} else if name != "" {
			username = name
		}

		attributes := make(map[string]string)
		for key, value := range userInfo {
			if key != "username" && key != "sub" {
				attributes[key] = value
			}
		}

		ctx.AuthenticatedUser = &authnmodel.AuthenticatedUser{
			IsAuthenticated:        true,
			UserID:                 sub,
			Username:               username,
			Domain:                 g.GetName(),
			AuthenticatedSubjectID: sub,
			Attributes:             attributes,
		}
	} else {
		// Fail the authentication if the authorization code is not provided
		ctx.AuthenticatedUser = &authnmodel.AuthenticatedUser{
			IsAuthenticated: false,
		}
	}

	// Set the flow response status based on the authentication result.
	if ctx.AuthenticatedUser.IsAuthenticated {
		execResp.Status = flowconst.ExecComplete
	} else {
		execResp.Status = flowconst.ExecUserError
		execResp.Type = flowconst.ExecRedirection
		execResp.Error = "User is not authenticated. Please provide a valid authorization code."
	}
}

// requiredInputData adds the required input data for the Google OIDC authentication flow.
// Returns true if input data should be requested from the user.
func (g *GoogleOIDCAuthExecutor) requiredInputData(ctx *flowmodel.FlowContext,
	execResp *flowmodel.ExecutorResponse) bool {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))

	// Check if the authorization code is already provided
	if code, ok := ctx.UserInputData["code"]; ok && code != "" {
		return false
	}

	// Define the authenticator specific required input data.
	googleReqData := []flowmodel.InputData{
		{
			Name:     "code",
			Type:     "string",
			Required: true,
		},
	}

	// Check for the required input data. Also appends the authenticator specific input data.
	requiredData := ctx.CurrentNode.GetInputData()
	if len(requiredData) == 0 {
		logger.Debug("No required input data defined for Google OIDC auth executor")
		// If no required input data is defined, use the default required data.
		requiredData = googleReqData
	} else {
		// Append the default required data if not already present.
		for _, inputData := range googleReqData {
			exists := false
			for _, existingInputData := range requiredData {
				if existingInputData.Name == inputData.Name {
					exists = true
					break
				}
			}
			// If the input data already exists, skip adding it again.
			if !exists {
				requiredData = append(requiredData, inputData)
			}
		}
	}

	requireData := false

	if execResp.RequiredData == nil {
		execResp.RequiredData = make([]flowmodel.InputData, 0)
	}

	if len(ctx.UserInputData) == 0 {
		execResp.RequiredData = append(execResp.RequiredData, requiredData...)
		return true
	}

	// Check if the required input data is provided by the user.
	for _, inputData := range requiredData {
		if _, ok := ctx.UserInputData[inputData.Name]; !ok {
			if !inputData.Required {
				logger.Debug("Skipping optional input data that is not provided by user",
					log.String("inputDataName", inputData.Name))
				continue
			}
			execResp.RequiredData = append(execResp.RequiredData, inputData)
			requireData = true
			logger.Debug("Required input data not provided by user", log.String("inputDataName", inputData.Name))
		}
	}

	return requireData
}

// GetUserInfo fetches user information from the Google OIDC provider using the access token.
func (g *GoogleOIDCAuthExecutor) GetUserInfo(ctx *flowmodel.FlowContext,
	accessToken string) (map[string]string, error) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, loggerComponentName))
	logger.Debug("Fetching user info from Google OIDC provider",
		log.String("userInfoEndpoint", g.GetUserInfoEndpoint()))

	// Create HTTP request
	req, err := http.NewRequest("GET", g.GetUserInfoEndpoint(), nil)
	if err != nil {
		logger.Error("Failed to create userinfo request", log.Error(err))
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}
	req.Header.Set(constants.AuthorizationHeaderName, constants.TokenTypeBearer+" "+accessToken)
	req.Header.Set(constants.AcceptHeaderName, "application/json")

	// Execute the request
	logger.Debug("Sending userinfo request to Google OIDC provider")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to send userinfo request", log.Error(err))
		return nil, fmt.Errorf("failed to send userinfo request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Failed to close userinfo response body", log.Error(closeErr))
		}
	}()
	logger.Debug("Userinfo response received from Google OIDC provider",
		log.Int("statusCode", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Userinfo request failed", log.String("status", resp.Status), log.String("body", string(body)))
		return nil, fmt.Errorf("userinfo request failed with status %s: %s", resp.Status, string(body))
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read userinfo response", log.Error(err))
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(body, &userInfo); err != nil {
		logger.Error("Failed to parse userinfo response", log.Error(err))
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	return systemutils.ConvertInterfaceMapToStringMap(userInfo), nil
}
