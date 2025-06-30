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

/*
MockGoogleIDP Usage Example:

To use the mock Google IDP in your tests:

1. Create a new mock IDP instance:
   mockIDP, err := testutils.NewMockGoogleIDP()
   if err != nil {
       t.Fatalf("Failed to create mock Google IDP: %v", err)
   }
   defer mockIDP.Close()

2. Configure your application to use the mock IDP endpoints:
   - Authorization URL: mockIDP.GetAuthorizationURL()
   - Token URL: mockIDP.GetTokenURL()
   - UserInfo URL: mockIDP.GetUserInfoURL()
   - JWKS URL: mockIDP.GetJWKSURL()

3. The mock IDP provides the following endpoints that mimic Google's OAuth2/OIDC behavior:
   - /o/oauth2/v2/auth - Authorization endpoint (returns hardcoded auth code)
   - /token - Token endpoint (exchanges auth code for access token and ID token)
   - /v1/userinfo - UserInfo endpoint (returns mock user information)
   - /oauth2/v3/certs - JWKS endpoint (returns public keys for token verification)

4. Pre-configured test data:
   - Authorization Code: "test_auth_code_12345"
   - User ID: "1234567890"
   - Email: "testuser@example.com"
   - Name: "Test User"

5. All tokens are properly signed with RSA256 and can be verified using the JWKS endpoint.
*/

package testutils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// MockGoogleIDP represents a mock Google Identity Provider server for testing
type MockGoogleIDP struct {
	Server       *httptest.Server
	PrivateKey   *rsa.PrivateKey
	PublicKey    *rsa.PublicKey
	KeyID        string
	ClientID     string
	ClientSecret string
	// Authorized codes to access tokens mapping
	AuthorizedCodes map[string]string
	// Access tokens to user info mapping
	AccessTokens map[string]MockUserInfo
}

// MockUserInfo represents mock user information returned by the IDP
type MockUserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// TokenResponse represents the OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token,omitempty"`
	Scope       string `json:"scope,omitempty"`
}

// NewMockGoogleIDP creates a new mock Google IDP server
func NewMockGoogleIDP() (*MockGoogleIDP, error) {
	// Generate RSA key pair for JWT signing
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	publicKey := &privateKey.PublicKey
	keyID := "test-key-id-" + strconv.FormatInt(time.Now().Unix(), 10)

	mockIDP := &MockGoogleIDP{
		PrivateKey:      privateKey,
		PublicKey:       publicKey,
		KeyID:           keyID,
		ClientID:        "test-google-client-id",
		ClientSecret:    "test_google_client_secret",
		AuthorizedCodes: make(map[string]string),
		AccessTokens:    make(map[string]MockUserInfo),
	}

	// Create the test server with HTTP handlers
	mux := http.NewServeMux()

	// OAuth2 Authorization endpoint
	mux.HandleFunc("/o/oauth2/v2/auth", mockIDP.handleAuthorize)

	// OAuth2 Token endpoint
	mux.HandleFunc("/token", mockIDP.handleToken)

	// OpenID Connect UserInfo endpoint
	mux.HandleFunc("/v1/userinfo", mockIDP.handleUserInfo)

	// JWKS endpoint
	mux.HandleFunc("/oauth2/v3/certs", mockIDP.handleJWKS)

	mockIDP.Server = httptest.NewServer(mux)

	// Pre-populate with test data
	mockIDP.setupTestData()

	log.Printf("Mock Google IDP server started at: %s", mockIDP.Server.URL)
	return mockIDP, nil
}

// Close shuts down the mock server
func (m *MockGoogleIDP) Close() {
	if m.Server != nil {
		m.Server.Close()
	}
}

// GetAuthorizationURL returns the authorization URL for the mock IDP
func (m *MockGoogleIDP) GetAuthorizationURL() string {
	return m.Server.URL + "/o/oauth2/v2/auth"
}

// GetTokenURL returns the token URL for the mock IDP
func (m *MockGoogleIDP) GetTokenURL() string {
	return m.Server.URL + "/token"
}

// GetUserInfoURL returns the user info URL for the mock IDP
func (m *MockGoogleIDP) GetUserInfoURL() string {
	return m.Server.URL + "/v1/userinfo"
}

// GetJWKSURL returns the JWKS URL for the mock IDP
func (m *MockGoogleIDP) GetJWKSURL() string {
	return m.Server.URL + "/oauth2/v3/certs"
}

// setupTestData pre-populates the mock IDP with test data
func (m *MockGoogleIDP) setupTestData() {
	// Add a test authorization code and corresponding access token
	authCode := "test_auth_code_12345"
	accessToken := "test_access_token_" + strconv.FormatInt(time.Now().Unix(), 10)

	m.AuthorizedCodes[authCode] = accessToken
	m.AccessTokens[accessToken] = MockUserInfo{
		Sub:           "1234567890",
		Name:          "Test User",
		GivenName:     "Test",
		FamilyName:    "User",
		Email:         "testuser@example.com",
		EmailVerified: true,
		Picture:       "https://example.com/avatar.jpg",
		Locale:        "en",
	}
}

// handleAuthorize handles the OAuth2 authorization endpoint
func (m *MockGoogleIDP) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock Google IDP: Authorization request received")

	// Parse query parameters
	clientID := r.URL.Query().Get("client_id")
	redirectURI := r.URL.Query().Get("redirect_uri")
	responseType := r.URL.Query().Get("response_type")
	state := r.URL.Query().Get("state")
	scope := r.URL.Query().Get("scope")

	// Validate basic parameters
	if clientID == "" || redirectURI == "" || responseType != "code" {
		http.Error(w, "Invalid request parameters", http.StatusBadRequest)
		return
	}

	log.Printf("Mock Google IDP: Redirecting with authorization code")

	// Generate authorization code and redirect back
	authCode := "test_auth_code_12345"

	// Create access token for this authorization code
	accessToken := "test_access_token_" + strconv.FormatInt(time.Now().Unix(), 10)
	m.AuthorizedCodes[authCode] = accessToken

	// Add user info for this access token if not exists
	if _, exists := m.AccessTokens[accessToken]; !exists {
		m.AccessTokens[accessToken] = MockUserInfo{
			Sub:           "1234567890",
			Name:          "Test User",
			GivenName:     "Test",
			FamilyName:    "User",
			Email:         "testuser@example.com",
			EmailVerified: true,
			Picture:       "https://example.com/avatar.jpg",
			Locale:        "en",
		}
	}

	// Build redirect URL with authorization code
	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "Invalid redirect URI", http.StatusBadRequest)
		return
	}

	params := url.Values{}
	params.Add("code", authCode)
	if state != "" {
		params.Add("state", state)
	}
	params.Add("scope", scope)

	redirectURL.RawQuery = params.Encode()

	// Send redirect response
	w.Header().Set("Location", redirectURL.String())
	w.WriteHeader(http.StatusFound)
}

// handleToken handles the OAuth2 token endpoint
func (m *MockGoogleIDP) handleToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock Google IDP: Token request received")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	// Validate grant type
	if grantType != "authorization_code" {
		http.Error(w, "Unsupported grant type", http.StatusBadRequest)
		return
	}

	// Basic validation (in a real implementation, you'd validate client credentials)
	if clientID == "" || code == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Log for debugging
	log.Printf("Token request - ClientID: %s, RedirectURI: %s", clientID, redirectURI)
	_ = clientSecret // Suppressing unused variable warning - would be used for validation in real implementation

	// Validate authorization code
	accessToken, exists := m.AuthorizedCodes[code]
	if !exists {
		http.Error(w, "Invalid authorization code", http.StatusBadRequest)
		return
	}

	// Generate ID token
	idToken, err := m.generateIDToken(accessToken)
	if err != nil {
		log.Printf("Failed to generate ID token: %v", err)
		http.Error(w, "Failed to generate ID token", http.StatusInternalServerError)
		return
	}

	// Create token response
	tokenResp := TokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   3600,
		IDToken:     idToken,
		Scope:       "openid email profile",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	if err := json.NewEncoder(w).Encode(tokenResp); err != nil {
		log.Printf("Failed to encode token response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Mock Google IDP: Token response sent successfully")
}

// handleUserInfo handles the OpenID Connect UserInfo endpoint
func (m *MockGoogleIDP) handleUserInfo(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock Google IDP: UserInfo request received")

	// Extract access token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
		return
	}

	accessToken := parts[1]

	// Look up user info for this access token
	userInfo, exists := m.AccessTokens[accessToken]
	if !exists {
		http.Error(w, "Invalid access token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userInfo); err != nil {
		log.Printf("Failed to encode user info response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Mock Google IDP: UserInfo response sent successfully")
}

// handleJWKS handles the JWKS endpoint
func (m *MockGoogleIDP) handleJWKS(w http.ResponseWriter, r *http.Request) {
	log.Printf("Mock Google IDP: JWKS request received")

	// Create JWK from RSA public key
	n := base64.RawURLEncoding.EncodeToString(m.PublicKey.N.Bytes())
	eBytes := []byte{1, 0, 1} // 65537 in big-endian
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": m.KeyID,
		"n":   n,
		"e":   e,
	}

	jwks := map[string]interface{}{
		"keys": []interface{}{jwk},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		log.Printf("Failed to encode JWKS response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Mock Google IDP: JWKS response sent successfully")
}

// generateIDToken creates a mock ID token
func (m *MockGoogleIDP) generateIDToken(accessToken string) (string, error) {
	userInfo, exists := m.AccessTokens[accessToken]
	if !exists {
		return "", fmt.Errorf("no user info found for access token")
	}

	now := time.Now()

	// Create JWT header
	header := map[string]interface{}{
		"alg": "RS256",
		"typ": "JWT",
		"kid": m.KeyID,
	}

	// Create JWT payload
	payload := map[string]interface{}{
		"iss":            m.Server.URL,
		"sub":            userInfo.Sub,
		"aud":            m.ClientID,
		"exp":            now.Add(time.Hour).Unix(),
		"iat":            now.Unix(),
		"auth_time":      now.Unix(),
		"email":          userInfo.Email,
		"email_verified": userInfo.EmailVerified,
		"name":           userInfo.Name,
		"given_name":     userInfo.GivenName,
		"family_name":    userInfo.FamilyName,
		"picture":        userInfo.Picture,
		"locale":         userInfo.Locale,
	}

	// Encode header and payload
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("failed to marshal header: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	headerBase64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadBase64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signing input and sign
	signingInput := headerBase64 + "." + payloadBase64
	hashed := sha256.Sum256([]byte(signingInput))

	signature, err := rsa.SignPKCS1v15(rand.Reader, m.PrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	signatureBase64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureBase64, nil
}

// AddUser adds a new user to the mock IDP for testing
func (m *MockGoogleIDP) AddUser(authCode, accessToken string, userInfo MockUserInfo) {
	m.AuthorizedCodes[authCode] = accessToken
	m.AccessTokens[accessToken] = userInfo
}

// GetClientID returns the mock client ID
func (m *MockGoogleIDP) GetClientID() string {
	return m.ClientID
}
