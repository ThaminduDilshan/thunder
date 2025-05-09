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

package model

import (
	"time"

	oauthmodel "github.com/asgardeo/thunder/internal/oauth/oauth2/model"
)

// TODO: Temporary adding the AuthenticatedUser struct here. Should be moved to a appropriate place.
type AuthenticatedUser struct {
	IsAuthenticated        bool
	UserId                 string
	Username               string
	Domain                 string
	AuthenticatedSubjectId string
	Attributes             map[string]string
}

type SessionData struct {
	OAuthParameters oauthmodel.OAuthParameters
	LoggedInUser    AuthenticatedUser
	AuthTime        time.Time
}
