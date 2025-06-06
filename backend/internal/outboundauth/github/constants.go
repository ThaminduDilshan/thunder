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

package github

// Constants used for GitHub OAuth2 authentication and API endpoints.
const (
	githubAuthorizeEndpoint = "https://github.com/login/oauth/authorize"
	githubTokenEndpoint     = "https://github.com/login/oauth/access_token" // #nosec G101
	githubUserInfoEndpoint  = "https://api.github.com/user"
	githubUserEmailEndpoint = "https://api.github.com/user/emails"
	userScope               = "user"
	userEmailScope          = "user:email"
)

// loggerComponentName is the name of the logger component for GitHub authenticator.
const loggerComponentName = "GithubAuthenticator"
