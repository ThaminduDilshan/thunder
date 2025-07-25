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

// Package dto defines the common data transfer objects for authentication handling.
package dto

import (
	"context"
	"time"
)

// AuthenticatedUser represents the user information of an authenticated user.
type AuthenticatedUser struct {
	IsAuthenticated bool
	UserID          string
	Attributes      map[string]string
}

// AuthenticationContext represents the context of an authentication session.
type AuthenticationContext struct {
	context.Context
	SessionDataKey     string
	RequestQueryParams map[string]string
	AuthenticatedUser  AuthenticatedUser
	AuthTime           time.Time
}
