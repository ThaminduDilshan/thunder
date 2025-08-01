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

// Package token provides handler for managing OAuth 2.0 token requests.
package token

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/asgardeo/thunder/internal/system/utils"

	appprovider "github.com/asgardeo/thunder/internal/application/provider"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/constants"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/granthandlers"
	"github.com/asgardeo/thunder/internal/oauth/oauth2/model"
	scopeprovider "github.com/asgardeo/thunder/internal/oauth/scope/provider"
	"github.com/asgardeo/thunder/internal/system/log"
)

// TokenHandlerInterface defines the interface for handling OAuth 2.0 token requests.
type TokenHandlerInterface interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request)
}

// TokenHandler implements the TokenHandlerInterface.
type TokenHandler struct {
	ApplicationProvider    appprovider.ApplicationProviderInterface
	ScopeValidatorProvider scopeprovider.ScopeValidatorProviderInterface
}

// NewTokenHandler creates a new instance of TokenHandler.
func NewTokenHandler() TokenHandlerInterface {
	return &TokenHandler{
		ApplicationProvider:    appprovider.NewApplicationProvider(),
		ScopeValidatorProvider: scopeprovider.NewScopeValidatorProvider(),
	}
}

// HandleTokenRequest handles the token request for OAuth 2.0.
// It validates the client credentials and delegates to the appropriate grant handler.
func (th *TokenHandler) HandleTokenRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, "TokenHandler"))

	// Parse the form data from the request body.
	if err := r.ParseForm(); err != nil {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Failed to parse request body", http.StatusBadRequest, nil)
		return
	}

	// Validate the grant_type.
	grantType := r.FormValue(constants.GrantType)
	if grantType == "" {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest,
			"Missing grant_type parameter", http.StatusBadRequest, nil)
		return
	}

	var grantHandler granthandlers.GrantHandler
	switch grantType {
	case constants.GrantTypeClientCredentials:
		grantHandler = &granthandlers.ClientCredentialsGrantHandler{}
	case constants.GrantTypeAuthorizationCode:
		grantHandler = &granthandlers.AuthorizationCodeGrantHandler{}
	case constants.GrantTypeRefreshToken:
		grantHandler = &granthandlers.RefreshTokenGrantHandler{}
	default:
		utils.WriteJSONError(w, constants.ErrorUnsupportedGrantType,
			"Unsupported grant type", http.StatusBadRequest, nil)
		return
	}

	// Extract client credentials from the request.
	clientID := ""
	clientSecret := ""
	if r.Header.Get("Authorization") != "" {
		var err error
		clientID, clientSecret, err = utils.ExtractBasicAuthCredentials(r)
		if err != nil {
			if err.Error() == "invalid authorization header" {
				responseHeaders := []map[string]string{
					{"WWW-Authenticate": "Basic"},
				}
				utils.WriteJSONError(w, constants.ErrorInvalidClient,
					"Invalid client credentials", http.StatusUnauthorized, responseHeaders)
				return
			}
			utils.WriteJSONError(w, constants.ErrorInvalidClient,
				"Invalid client credentials", http.StatusUnauthorized, nil)
			return
		}
	}

	// Check for client credentials in the request body.
	clientIDFromBody := r.FormValue(constants.ClientID)
	clientSecretFromBody := r.FormValue(constants.ClientSecret)

	if clientIDFromBody != "" && clientSecretFromBody != "" {
		if clientID != "" && clientSecret != "" {
			utils.WriteJSONError(w, constants.ErrorInvalidRequest,
				"Authorization information is provided in both header and body", http.StatusBadRequest, nil)
			return
		}

		clientID = clientIDFromBody
		clientSecret = clientSecretFromBody
	} else {
		if clientID == "" {
			clientID = clientIDFromBody
		}
		if clientSecret == "" {
			clientSecret = clientSecretFromBody
		}
	}

	if clientID == "" {
		utils.WriteJSONError(w, constants.ErrorInvalidRequest, "Missing client_id parameter",
			http.StatusBadRequest, nil)
		return
	}

	// Retrieve the OAuth application based on the client id.
	appService := th.ApplicationProvider.GetApplicationService()
	oauthApp, err := appService.GetOAuthApplication(clientID)
	if err != nil || oauthApp == nil {
		utils.WriteJSONError(w, constants.ErrorInvalidClient,
			"Invalid client credentials", http.StatusUnauthorized, nil)
		return
	}

	// Validate grant type against the application.
	if !oauthApp.IsAllowedGrantType(grantType) {
		utils.WriteJSONError(w, constants.ErrorUnauthorizedClient,
			"The client is not authorized to use this grant type", http.StatusUnauthorized, nil)
		return
	}

	// Construct the token request.
	tokenRequest := &model.TokenRequest{
		GrantType:    grantType,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        r.FormValue("scope"),
		Username:     r.FormValue("username"),
		Password:     r.FormValue("password"),
		RefreshToken: r.FormValue("refresh_token"),
		CodeVerifier: r.FormValue("code_verifier"),
		Code:         r.FormValue("code"),
		RedirectURI:  r.FormValue("redirect_uri"),
	}

	// Validate the token request.
	tokenError := grantHandler.ValidateGrant(tokenRequest, oauthApp)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Validate and filter scopes.
	scopeValidator := th.ScopeValidatorProvider.GetScopeValidator()
	validScopes, scopeError := scopeValidator.ValidateScopes(tokenRequest.Scope, oauthApp.ClientID)
	if scopeError != nil {
		utils.WriteJSONError(w, scopeError.Error, scopeError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}
	tokenRequest.Scope = validScopes

	// Delegate to the grant handler.
	ctx := &model.TokenContext{
		TokenAttributes: make(map[string]interface{}),
	}
	tokenRespDTO, tokenError := grantHandler.HandleGrant(tokenRequest, oauthApp, ctx)
	if tokenError != nil && tokenError.Error != "" {
		utils.WriteJSONError(w, tokenError.Error, tokenError.ErrorDescription, http.StatusBadRequest, nil)
		return
	}

	// Generate and add refresh token if applicable.
	if tokenRequest.GrantType == constants.GrantTypeAuthorizationCode &&
		oauthApp.IsAllowedGrantType(constants.GrantTypeRefreshToken) {
		logger.Debug("Issuing refresh token for the token request", log.String("client_id", clientID),
			log.String("grant_type", grantType))

		refreshGrantHandler := &granthandlers.RefreshTokenGrantHandler{}
		refreshTokenError := refreshGrantHandler.IssueRefreshToken(tokenRespDTO, ctx, oauthApp.ClientID,
			grantType, tokenRespDTO.AccessToken.Scopes)
		if refreshTokenError != nil && refreshTokenError.Error != "" {
			utils.WriteJSONError(w, refreshTokenError.Error, refreshTokenError.ErrorDescription,
				http.StatusInternalServerError, nil)
			return
		}
	}

	scopes := strings.Join(tokenRespDTO.AccessToken.Scopes, " ")
	tokenResponse := &model.TokenResponse{
		AccessToken:  tokenRespDTO.AccessToken.Token,
		TokenType:    tokenRespDTO.AccessToken.TokenType,
		ExpiresIn:    tokenRespDTO.AccessToken.ExpiresIn,
		RefreshToken: tokenRespDTO.RefreshToken.Token,
		Scope:        scopes,
	}

	logger.Debug("Token generated successfully", log.String("client_id", clientID),
		log.String("grant_type", grantType))

	// Set the response headers.
	w.Header().Set("Content-Type", "application/json")
	// Must include the following headers when sensitive data is returned.
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	// Write the token response.
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tokenResponse); err != nil {
		logger.Error("Failed to write token response", log.Error(err))
		http.Error(w, "Failed to write token response", http.StatusInternalServerError)
		return
	}
	logger.Debug("Token response sent", log.String("client_id", clientID), log.String("grant_type", grantType))
}
