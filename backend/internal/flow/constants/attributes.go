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

package constants

// Attribute constants for user attributes used in the flow.
const AttributeUserID = "userID"
const AttributeUsername = "username"
const AttributePassword = "password"
const AttributeEmail = "email"
const AttributeMobileNumber = "mobileNumber"
const AttributeFirstName = "firstName"
const AttributeLastName = "lastName"

// UserIdentifiableAttributes contains the list of user attributes that can be used to identify a user.
var UserIdentifiableAttributes = []string{
	AttributeUserID,
	AttributeUsername,
	AttributeEmail,
	AttributeMobileNumber,
	AttributeFirstName,
	AttributeLastName,
}

// UserCredentialAttributes contains the list of user attributes that are considered as credentials.
var UserCredentialAttributes = []string{
	AttributePassword,
}
