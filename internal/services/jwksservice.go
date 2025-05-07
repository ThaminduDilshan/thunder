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

package services

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"net/http"

	"github.com/asgardeo/thunder/internal/cert"
	oauthutils "github.com/asgardeo/thunder/internal/identity/oauth2/utils"
	"github.com/asgardeo/thunder/internal/server"
	"github.com/asgardeo/thunder/internal/system/config"
)

// JWKSService for handling JWKS requests.
type JWKSService struct{}

func NewJWKSService(mux *http.ServeMux) *JWKSService {

	jwksService := &JWKSService{}
	jwksService.RegisterRoutes(mux)
	return jwksService
}

func (s *JWKSService) RegisterRoutes(mux *http.ServeMux) {

	mux.HandleFunc("GET /oauth2/jwks", s.HandleJWKS)
}

// HandleJWKS is the handler function for the JWKS endpoint.
func (s *JWKSService) HandleJWKS(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// Add CORS headers if the request is from a allowed origin.
	allowedOrigins, err := server.GetAllowedOrigins()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Something went wrong"}`))
		return
	}
	allowedOrigin := oauthutils.GetAllowedOrigin(allowedOrigins, r.Header.Get("Origin"))
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}

	// Load configuration and current directory.
	thunderRuntime := config.GetThunderRuntime()

	// Retrieve the TLS configuration.
	tlsConfig, err := cert.GetTLSConfig(&thunderRuntime.Config, thunderRuntime.ThunderHome)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Unable to load TLS configuration"}`))
		return
	}

	// Extract the public key from the certificate.
	certData := tlsConfig.Certificates[0].Certificate[0]
	parsedCert, err := x509.ParseCertificate(certData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Unable to parse certificate"}`))
		return
	}

	// Convert the public key to JWKS format.
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"n":   parsedCert.PublicKey.(*rsa.PublicKey).N.String(),
				"e":   "AQAB",
			},
		},
	}

	// Serialize JWKS to JSON.
	jwksJSON, err := json.Marshal(jwks)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Unable to serialize JWKS"}`))
		return
	}

	// Write the JWKS response.
	w.WriteHeader(http.StatusOK)
	w.Write(jwksJSON)
}
