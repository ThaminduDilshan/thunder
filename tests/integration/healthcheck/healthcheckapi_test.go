/*
 * Copyright (c) 2025, WSO2 LLC. (http://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
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

package healthcheck

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	testServerURL = "https://localhost:8095"
)

type HealthCheckAPITestSuite struct {
	suite.Suite
}

func TestHealthCheckAPITestSuite(t *testing.T) {
	suite.Run(t, new(HealthCheckAPITestSuite))
}

// TestLivenessCheck tests the liveness endpoint.
func (ts *HealthCheckAPITestSuite) TestLivenessCheck() {
	req, err := http.NewRequest("GET", testServerURL+"/health/liveness", nil)
	ts.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, resp.StatusCode)
}

// TestReadinessCheck tests the readiness endpoint.
func (ts *HealthCheckAPITestSuite) TestReadinessCheck() {
	req, err := http.NewRequest("GET", testServerURL+"/health/readiness", nil)
	ts.Require().NoError(err)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	ts.Require().NoError(err)
	ts.Require().Equal(http.StatusOK, resp.StatusCode)

	var healthStatus map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthStatus)
	ts.Require().NoError(err)
	ts.Require().Equal("UP", healthStatus["status"])
	ts.Require().NotEmpty(healthStatus["service_status"])
}
