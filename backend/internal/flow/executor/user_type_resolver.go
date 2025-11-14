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

import (
	"fmt"
	"slices"

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
	"github.com/asgardeo/thunder/internal/userschema"
)

const (
	userTypeResolverLoggerComponentName = "UserTypeResolver"
)

// userTypeResolver is a registration-flow executor that resolves the user type at flow start.
type userTypeResolver struct {
	flowcore.ExecutorInterface
	userSchemaService userschema.UserSchemaServiceInterface
	logger            *log.Logger
}

var _ flowcore.ExecutorInterface = (*userTypeResolver)(nil)

// newUserTypeResolver creates a new instance of the UserTypeResolver executor.
func newUserTypeResolver(
	flowFactory flowcore.FlowFactoryInterface,
	userSchemaService userschema.UserSchemaServiceInterface,
) *userTypeResolver {
	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, userTypeResolverLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameUserTypeResolver))

	base := flowFactory.CreateExecutor(ExecutorNameUserTypeResolver, flowcm.ExecutorTypeRegistration,
		[]flowcm.InputData{}, []flowcm.InputData{})

	return &userTypeResolver{
		ExecutorInterface: base,
		userSchemaService: userSchemaService,
		logger:            logger,
	}
}

// Execute resolves the user type from inputs or prompts the user to select one.
func (u *userTypeResolver) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := u.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing user type resolver")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	// Only applies to registration flows
	if ctx.FlowType != flowcm.FlowTypeRegistration {
		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	allowed := ctx.Application.AllowedUserTypes

	// If a userType is provided in inputs, validate and accept it
	if userType, ok := ctx.UserInputData[inputUserType]; ok && userType != "" {
		if len(allowed) == 0 || slices.Contains(allowed, userType) {
			logger.Debug("User type resolved from input", log.String("userType", userType))

			// Check if self registration is enabled for the user type
			userSchema, ouID, svcErr := u.getUserSchemaAndOU(userType, logger)
			if svcErr != nil {
				return execResp, fmt.Errorf("failed to resolve user schema for user type")
			}

			if !userSchema.AllowSelfRegistration {
				logger.Debug("Self registration not enabled for user type", log.String("userType", userType))
				execResp.Status = flowcm.ExecFailure
				execResp.FailureReason = fmt.Sprintf("Self-registration not enabled for user type: %s", userType)
				return execResp, nil
			}

			// Add userType to runtime data
			execResp.RuntimeData[inputUserType] = userType
			execResp.RuntimeData[defaultOUIDKey] = ouID

			execResp.Status = flowcm.ExecComplete
			return execResp, nil
		}
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = fmt.Sprintf("Registration not allowed for user type: %s", userType)
		return execResp, nil
	}

	// Check for allowed user types to decide next steps
	if len(allowed) == 0 {
		// TODO: This should be improved to fallback to the application's ou when the support is available.
		//  userType has an attached ou. Need to find userType from the application's ou.
		//  Also should check if self registration is enabled for the user type when the support is available.

		logger.Debug("No allowed user types found for the application")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Self-registration not available for this application"
		return execResp, nil
	}

	if len(allowed) == 1 {
		// Check if self registration is enabled for the user type
		userSchema, ouID, svcErr := u.getUserSchemaAndOU(allowed[0], logger)
		if svcErr != nil {
			return execResp, fmt.Errorf("failed to resolve user schema for user type")
		}

		if !userSchema.AllowSelfRegistration {
			logger.Debug("Self registration not enabled for user type", log.String("userType", allowed[0]))
			execResp.Status = flowcm.ExecFailure
			execResp.FailureReason = "Self-registration not available for this application"
			return execResp, nil
		}

		logger.Debug("User type resolved from allowed list", log.String("userType", allowed[0]))

		// Add userType to runtime data
		execResp.RuntimeData[inputUserType] = allowed[0]
		execResp.RuntimeData[defaultOUIDKey] = ouID

		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	// Filter user types to only those with self registration enabled
	var selfRegUserTypes []string
	for _, userType := range allowed {
		userSchema, _, svcErr := u.getUserSchemaAndOU(userType, logger)
		if svcErr != nil {
			logger.Debug("Failed to resolve user schema", log.String("userType", userType))
			continue
		}
		if userSchema.AllowSelfRegistration {
			selfRegUserTypes = append(selfRegUserTypes, userType)
		}
	}

	// If no user types have self registration enabled
	if len(selfRegUserTypes) == 0 {
		logger.Debug("No user types with self registration enabled")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Self-registration not available for this application"
		return execResp, nil
	}

	// If only one user type has self registration enabled, select it automatically
	if len(selfRegUserTypes) == 1 {
		_, ouID, svcErr := u.getUserSchemaAndOU(selfRegUserTypes[0], logger)
		if svcErr != nil {
			return execResp, fmt.Errorf("failed to resolve user schema for user type")
		}

		logger.Debug("User type auto-selected", log.String("userType", selfRegUserTypes[0]))
		execResp.RuntimeData[inputUserType] = selfRegUserTypes[0]
		execResp.RuntimeData[defaultOUIDKey] = ouID

		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	// If multiple user types are allowed, prompt the user to select one
	logger.Debug("Prompting for user type selection", log.Any("userTypes", selfRegUserTypes))

	execResp.Status = flowcm.ExecUserInputRequired
	execResp.RequiredData = []flowcm.InputData{
		{
			Name:     inputUserType,
			Type:     "dropdown",
			Required: true,
			Options:  append([]string{}, selfRegUserTypes...),
		},
	}
	return execResp, nil
}

// getUserSchemaAndOU retrieves the user schema by name and returns the schema and organization unit ID.
func (u *userTypeResolver) getUserSchemaAndOU(userType string, logger *log.Logger) (*userschema.UserSchema, string, *serviceerror.ServiceError) {
	userSchema, svcErr := u.userSchemaService.GetUserSchemaByName(userType)
	if svcErr != nil {
		logger.Error("Failed to resolve user schema for user type",
			log.String("userType", userType), log.String("error", svcErr.Error))
		return nil, "", svcErr
	}

	if userSchema.OrganizationUnitID == "" {
		logger.Error("No organization unit found for user type", log.String("userType", userType))
		return nil, "", &serviceerror.ServiceError{
			Type:             serviceerror.ServerErrorType,
			Code:             "UT-500",
			Error:            "Internal Server Error",
			ErrorDescription: "No organization unit found for user type",
		}
	}

	logger.Debug("User schema resolved for user type",
		log.String("userType", userType), log.String("ouID", userSchema.OrganizationUnitID))
	return userSchema, userSchema.OrganizationUnitID, nil
}
