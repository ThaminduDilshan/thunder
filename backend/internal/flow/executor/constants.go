/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
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

package executor

// Executor name constants
const (
	ExecutorNameBasicAuth        = "BasicAuthExecutor"
	ExecutorNameSMSAuth          = "SMSOTPAuthExecutor"
	ExecutorNameOAuth            = "OAuthExecutor"
	ExecutorNameOIDCAuth         = "OIDCAuthExecutor"
	ExecutorNameGitHubAuth       = "GithubOAuthExecutor"
	ExecutorNameGoogleAuth       = "GoogleOIDCAuthExecutor"
	ExecutorNameIdentifying      = "IdentifyingExecutor"
	ExecutorNameAuthAssert       = "AuthAssertExecutor"
	ExecutorNameProvisioning     = "ProvisioningExecutor"
	ExecutorNameAttributeCollect = "AttributeCollector"
	ExecutorNameAuthorization    = "AuthorizationExecutor"
	ExecutorNameOUCreation       = "OUExecutor"
	ExecutorNameHTTPRequest      = "HTTPRequestExecutor"
	ExecutorNameUserTypeResolver = "UserTypeResolver"
)

// User attribute and input constants
const (
	userAttributeUsername     = "username"
	userAttributePassword     = "password"
	userAttributeUserID       = "userID"
	userAttributeMobileNumber = "mobileNumber"
	userAttributeEmail        = "email"
	userAttributeGroups       = "groups"

	userInputOuName   = "ouName"
	userInputOuHandle = "ouHandle"
	userInputOuDesc   = "ouDescription"
	ouIDKey           = "ouId"
	defaultOUIDKey    = "defaultOUID"

	inputUserType = "userType"
)

// nonSearchableInputs contains the list of user inputs/ attributes that are non-searchable.
var nonSearchableInputs = []string{"password", "code", "nonce", "otp"}

// nonUserAttributes contains the list of user attributes that do not belong to user entity.
var nonUserAttributes = []string{"userID", "code", "nonce", "state", "flowID",
	"otp", "attemptCount", "expiryTimeInMillis", "value", "userType", "ouId", "defaultOUID"}

// Failure reason constants
const (
	failureReasonUserNotAuthenticated     = "User is not authenticated"
	failureReasonUserNotFound             = "User not found"
	failureReasonInvalidAuthorizationCode = "Authentication failed. Authorization code not provided or invalid"
)
