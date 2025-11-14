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
	"errors"
	"fmt"
	"slices"

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
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

			// TODO: Check if self registration is enabled for the user type when the support is available

			// Add userType to runtime data
			execResp.RuntimeData[inputUserType] = userType

			if err := u.resolveAndSetDefaultOUID(userType, execResp); err != nil {
				return execResp, err
			}

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
		// TODO: Check if self registration is enabled for the user type when the support is available

		logger.Debug("User type resolved from allowed list", log.String("userType", allowed[0]))

		// Add userType to runtime data
		execResp.RuntimeData[inputUserType] = allowed[0]

		if err := u.resolveAndSetDefaultOUID(allowed[0], execResp); err != nil {
			return execResp, err
		}

		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	// TODO: Should verify which user types have self registration enabled when the support is available.
	//  If only one user type remains after filtering, select it automatically.

	// If multiple user types are allowed, prompt the user to select one
	logger.Debug("Prompting for user type selection", log.Any("userTypes", allowed))

	execResp.Status = flowcm.ExecUserInputRequired
	execResp.RequiredData = []flowcm.InputData{
		{
			Name:     inputUserType,
			Type:     "dropdown",
			Required: true,
			Options:  append([]string{}, allowed...),
		},
	}
	return execResp, nil
}

// resolveAndSetDefaultOUID resolves the organization unit for the given user type and sets it in runtime data.
func (u *userTypeResolver) resolveAndSetDefaultOUID(userType string, execResp *flowcm.ExecutorResponse) error {
	logger := u.logger

	// TODO: Uncomment when the implementation is available
	// ouID, svcErr := u.userSchemaService.GetOUForUserType(userType)
	// if svcErr != nil {
	// 	logger.Error("Failed to resolve organization unit for user type",
	// 		log.String("userType", userType), log.String("error", svcErr.Error))
	// 	return fmt.Errorf("failed to resolve organization unit for user type")
	// }
	// if ouID != nil && *ouID != "" {
	// 	logger.Debug("Organization unit resolved for user type",
	// 		log.String("userType", userType), log.String("ouID", *ouID))
	// 	execResp.RuntimeData[defaultOUIDKey] = *ouID
	// 	return nil
	// }

	logger.Error("No organization unit found for user type", log.String("userType", userType))
	return errors.New("no organization unit found for user type")
}
