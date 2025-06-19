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

package server

// import (
// 	"errors"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/asgardeo/thunder/internal/system/database/model"
// 	dbprovider "github.com/asgardeo/thunder/internal/system/database/provider"
// 	"github.com/asgardeo/thunder/tests/mocks/databasemock"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/suite"
// )

// // ServerOperationTestSuite defines the test suite for server operations
// type ServerOperationTestSuite struct {
// 	suite.Suite
// 	mockDBProvider   *databasemock.MockDBProvider
// 	mockDBClient     *databasemock.MockDBClient
// 	originalProvider dbprovider.DBProviderInterface
// }

// func TestServerOperationSuite(t *testing.T) {
// 	suite.Run(t, new(ServerOperationTestSuite))
// }

// func (suite *ServerOperationTestSuite) SetupTest() {
// 	// Store the original provider to restore later
// 	suite.originalProvider = dbprovider.NewDBProvider()

// 	// Initialize mock DB provider and client
// 	suite.mockDBProvider = &databasemock.MockDBProvider{}
// 	suite.mockDBClient = &databasemock.MockDBClient{}

// 	// Configure the mock provider to return our mock client
// 	suite.mockDBProvider.MockGetDBClient = func(dbName string) (client, error) {
// 		if dbName == "identity" {
// 			return suite.mockDBClient, nil
// 		}
// 		return nil, errors.New("unsupported database name")
// 	}

// 	// Replace the DBProvider constructor with our mock
// 	dbprovider.NewDBProvider = func() dbprovider.DBProviderInterface {
// 		return suite.mockDBProvider
// 	}
// }

// func (suite *ServerOperationTestSuite) TearDownTest() {
// 	// Restore the original provider
// 	dbprovider.NewDBProvider = func() dbprovider.DBProviderInterface {
// 		return suite.originalProvider
// 	}
// }

// func (suite *ServerOperationTestSuite) TestGetAllowedOriginsSuccess() {
// 	// Prepare the mock client to return a valid response
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		assert.Equal(suite.T(), QueryAllowedOrigins.ID, query.ID)
// 		assert.Equal(suite.T(), QueryAllowedOrigins.Query, query.Query)

// 		return []map[string]interface{}{
// 			{
// 				"allowed_origins": "https://example.com,https://test.com",
// 			},
// 		}, nil
// 	}

// 	// Call the function
// 	origins, err := getAllowedOrigins()

// 	// Verify the results
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), []string{"https://example.com", "https://test.com"}, origins)
// 	assert.Len(suite.T(), suite.mockDBProvider.GetDBClientCalls, 1)
// 	assert.Equal(suite.T(), "identity", suite.mockDBProvider.GetDBClientCalls[0])
// 	assert.Len(suite.T(), suite.mockDBClient.QueryCalls, 1)
// }

// func (suite *ServerOperationTestSuite) TestGetAllowedOriginsEmpty() {
// 	// Prepare the mock client to return an empty result
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return []map[string]interface{}{}, nil
// 	}

// 	// Call the function
// 	origins, err := getAllowedOrigins()

// 	// Verify the results
// 	assert.NoError(suite.T(), err)
// 	assert.Empty(suite.T(), origins)
// }

// func (suite *ServerOperationTestSuite) TestGetAllowedOriginsDBError() {
// 	// Prepare the mock client to return an error
// 	expectedError := errors.New("database query failed")
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return nil, expectedError
// 	}

// 	// Call the function
// 	origins, err := getAllowedOrigins()

// 	// Verify the results
// 	assert.Error(suite.T(), err)
// 	assert.Equal(suite.T(), expectedError, err)
// 	assert.Nil(suite.T(), origins)
// }

// func (suite *ServerOperationTestSuite) TestGetAllowedOriginsDBClientError() {
// 	// Prepare the mock provider to return an error
// 	expectedError := errors.New("failed to get DB client")
// 	suite.mockDBProvider.MockGetDBClient = func(dbName string) (client, error) {
// 		return nil, expectedError
// 	}

// 	// Call the function
// 	origins, err := getAllowedOrigins()

// 	// Verify the results
// 	assert.Error(suite.T(), err)
// 	assert.Equal(suite.T(), expectedError, err)
// 	assert.Nil(suite.T(), origins)
// }

// func (suite *ServerOperationTestSuite) TestGetAllowedOriginsInvalidType() {
// 	// Prepare the mock client to return an invalid type for allowed_origins
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return []map[string]interface{}{
// 			{
// 				"allowed_origins": 12345, // Not a string
// 			},
// 		}, nil
// 	}

// 	// Call the function
// 	origins, err := getAllowedOrigins()

// 	// Verify the results
// 	assert.Error(suite.T(), err)
// 	assert.Nil(suite.T(), origins)
// }

// func (suite *ServerOperationTestSuite) TestAddAllowedOriginHeadersSuccess() {
// 	// Prepare the mock client to return valid origins
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return []map[string]interface{}{
// 			{
// 				"allowed_origins": "https://example.com,https://test.com",
// 			},
// 		}, nil
// 	}

// 	// Create a test request with an Origin header
// 	req := httptest.NewRequest("GET", "/api/test", nil)
// 	req.Header.Set("Origin", "https://example.com")

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Create options
// 	options := &RequestWrapOptions{
// 		Cors: CorsOptions{
// 			AllowedMethods:   "GET, POST, PUT, DELETE",
// 			AllowedHeaders:   "Content-Type, Authorization",
// 			AllowCredentials: true,
// 		},
// 	}

// 	// Call the function
// 	err := addAllowedOriginHeaders(rr, req, options)

// 	// Verify the results
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
// 	assert.Equal(suite.T(), "GET, POST, PUT, DELETE", rr.Header().Get("Access-Control-Allow-Methods"))
// 	assert.Equal(suite.T(), "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
// 	assert.Equal(suite.T(), "true", rr.Header().Get("Access-Control-Allow-Credentials"))
// }

// func (suite *ServerOperationTestSuite) TestAddAllowedOriginHeadersNoMatch() {
// 	// Prepare the mock client to return valid origins
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return []map[string]interface{}{
// 			{
// 				"allowed_origins": "https://example.com,https://test.com",
// 			},
// 		}, nil
// 	}

// 	// Create a test request with an Origin header that's not in the allowed list
// 	req := httptest.NewRequest("GET", "/api/test", nil)
// 	req.Header.Set("Origin", "https://malicious.com")

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Create options
// 	options := &RequestWrapOptions{
// 		Cors: CorsOptions{
// 			AllowedMethods:   "GET, POST",
// 			AllowedHeaders:   "Content-Type",
// 			AllowCredentials: true,
// 		},
// 	}

// 	// Call the function
// 	err := addAllowedOriginHeaders(rr, req, options)

// 	// Verify the results
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "", rr.Header().Get("Access-Control-Allow-Origin"))
// 	assert.Equal(suite.T(), "", rr.Header().Get("Access-Control-Allow-Methods"))
// }

// func (suite *ServerOperationTestSuite) TestAddAllowedOriginHeadersNoOriginHeader() {
// 	// Create a test request without an Origin header
// 	req := httptest.NewRequest("GET", "/api/test", nil)

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Create options
// 	options := &RequestWrapOptions{
// 		Cors: CorsOptions{
// 			AllowedMethods:   "GET, POST",
// 			AllowedHeaders:   "Content-Type",
// 			AllowCredentials: true,
// 		},
// 	}

// 	// Call the function
// 	err := addAllowedOriginHeaders(rr, req, options)

// 	// Verify the results
// 	assert.NoError(suite.T(), err)
// 	assert.Equal(suite.T(), "", rr.Header().Get("Access-Control-Allow-Origin"))
// }

// func (suite *ServerOperationTestSuite) TestAddAllowedOriginHeadersDBError() {
// 	// Prepare the mock client to return an error
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return nil, errors.New("database error")
// 	}

// 	// Create a test request with an Origin header
// 	req := httptest.NewRequest("GET", "/api/test", nil)
// 	req.Header.Set("Origin", "https://example.com")

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Create options
// 	options := &RequestWrapOptions{
// 		Cors: CorsOptions{},
// 	}

// 	// Call the function
// 	err := addAllowedOriginHeaders(rr, req, options)

// 	// Verify the results
// 	assert.Error(suite.T(), err)
// 	assert.Contains(suite.T(), err.Error(), "failed to get allowed origins")
// }

// func (suite *ServerOperationTestSuite) TestWrapHandleFunction() {
// 	// Create a test mux
// 	mux := http.NewServeMux()

// 	// Create a handler function
// 	handlerCalled := false
// 	handler := func(w http.ResponseWriter, r *http.Request) {
// 		handlerCalled = true
// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("OK"))
// 	}

// 	// Prepare the mock client to return valid origins
// 	suite.mockDBClient.MockQuery = func(query model.DBQuery, args ...interface{}) ([]map[string]interface{}, error) {
// 		return []map[string]interface{}{
// 			{
// 				"allowed_origins": "https://example.com",
// 			},
// 		}, nil
// 	}

// 	// Create options
// 	options := &RequestWrapOptions{
// 		Cors: CorsOptions{
// 			AllowedMethods:   "GET, POST",
// 			AllowedHeaders:   "Content-Type",
// 			AllowCredentials: true,
// 		},
// 	}

// 	// Wrap the handler function
// 	WrapHandleFunction(mux, "/api/test", options, handler)

// 	// Create a test request
// 	req := httptest.NewRequest("GET", "/api/test", nil)
// 	req.Header.Set("Origin", "https://example.com")

// 	// Create a response recorder
// 	rr := httptest.NewRecorder()

// 	// Serve the request
// 	mux.ServeHTTP(rr, req)

// 	// Verify the results
// 	assert.True(suite.T(), handlerCalled)
// 	assert.Equal(suite.T(), http.StatusOK, rr.Code)
// 	assert.Equal(suite.T(), "OK", rr.Body.String())
// 	assert.Equal(suite.T(), "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
// 	assert.Equal(suite.T(), "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
// 	assert.Equal(suite.T(), "Content-Type", rr.Header().Get("Access-Control-Allow-Headers"))
// 	assert.Equal(suite.T(), "true", rr.Header().Get("Access-Control-Allow-Credentials"))
// }

// // Interface alias for better readability in test code
// type client = interface{}
