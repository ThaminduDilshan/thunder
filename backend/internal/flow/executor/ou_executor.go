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

	flowcm "github.com/asgardeo/thunder/internal/flow/common"
	flowcore "github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/ou"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
	"github.com/asgardeo/thunder/internal/system/log"
)

const (
	ouExecLoggerComponentName = "OUExecutor"
)

// ouExecutor is responsible for creating organizational units (OUs) within the system.
type ouExecutor struct {
	flowcore.ExecutorInterface
	ouService ou.OrganizationUnitServiceInterface
	logger    *log.Logger
}

var _ flowcore.ExecutorInterface = (*ouExecutor)(nil)

// newOUExecutor creates a new instance of OUExecutor with the given parameters.
func newOUExecutor(
	flowFactory flowcore.FlowFactoryInterface,
	ouService ou.OrganizationUnitServiceInterface,
) *ouExecutor {
	defaultInputs := []flowcm.InputData{
		{
			Name:     userInputOuName,
			Required: true,
			Type:     "string",
		},
		{
			Name:     userInputOuHandle,
			Required: true,
			Type:     "string",
		},
		{
			Name:     userInputOuDesc,
			Required: false,
			Type:     "string",
		},
	}

	logger := log.GetLogger().With(log.String(log.LoggerKeyComponentName, ouExecLoggerComponentName),
		log.String(log.LoggerKeyExecutorName, ExecutorNameOUCreation))

	base := flowFactory.CreateExecutor(ExecutorNameOUCreation, flowcm.ExecutorTypeRegistration,
		defaultInputs, []flowcm.InputData{})

	return &ouExecutor{
		ExecutorInterface: base,
		ouService:         ouService,
		logger:            logger,
	}
}

// Execute executes the ou creation logic.
func (o *ouExecutor) Execute(ctx *flowcore.NodeContext) (*flowcm.ExecutorResponse, error) {
	logger := o.logger.With(log.String(log.LoggerKeyFlowID, ctx.FlowID))
	logger.Debug("Executing OU creation executor")

	execResp := &flowcm.ExecutorResponse{
		AdditionalData: make(map[string]string),
		RuntimeData:    make(map[string]string),
	}

	if ctx.FlowType != flowcm.FlowTypeRegistration {
		logger.Debug("Flow type is not registration, skipping OU creation")
		execResp.Status = flowcm.ExecComplete
		return execResp, nil
	}

	if !o.ValidatePrerequisites(ctx, execResp) {
		logger.Debug("Prerequisites validation failed for OU creation")
		execResp.Status = flowcm.ExecFailure
		execResp.FailureReason = "Prerequisites validation failed for OU creation"
		return execResp, nil
	}

	if o.CheckInputData(ctx, execResp) {
		logger.Debug("Required input data for OU creation is not provided")
		execResp.Status = flowcm.ExecUserInputRequired
		return execResp, nil
	}

	// Create the OU using the OU service.
	ouRequest := o.getOrganizationUnitRequest(ctx)
	createdOU, svcErr := o.ouService.CreateOrganizationUnit(ouRequest)
	if svcErr != nil {
		if svcErr.Type == serviceerror.ClientErrorType {
			execResp.Status = flowcm.ExecFailure

			switch svcErr.Code {
			case ou.ErrorOrganizationUnitNameConflict.Code:
				execResp.FailureReason = "An organization unit with the same name already exists."
			case ou.ErrorOrganizationUnitHandleConflict.Code:
				execResp.FailureReason = "An organization unit with the same handle already exists."
			default:
				execResp.FailureReason = "Failed to create organization unit: " + svcErr.ErrorDescription
			}

			return execResp, nil
		}

		logger.Error("Error occurred while creating organization unit: ", log.String("errorCode", svcErr.Code),
			log.String("errorDescription", svcErr.ErrorDescription))
		return nil, errors.New("failed to create organization unit")
	}

	if createdOU.ID == "" {
		logger.Error("Organization unit creation failed: received empty OU ID")
		return nil, errors.New("failed to create organization unit")
	}

	// Set the created OU ID in the runtime data for further use in the flow.
	execResp.RuntimeData[ouIDKey] = createdOU.ID

	logger.Debug("Organization unit created successfully", log.String(ouIDKey, createdOU.ID))
	execResp.Status = flowcm.ExecComplete
	return execResp, nil
}

// getOrganizationUnitRequest constructs an OrganizationUnitRequest from the NodeContext.
func (o *ouExecutor) getOrganizationUnitRequest(ctx *flowcore.NodeContext) ou.OrganizationUnitRequest {
	ouRequest := ou.OrganizationUnitRequest{
		Name:        ctx.UserInputData[userInputOuName],
		Handle:      ctx.UserInputData[userInputOuHandle],
		Description: ctx.UserInputData[userInputOuDesc],
	}

	// Set parent OU ID if defaultOUID is present in runtime data
	if val, ok := ctx.RuntimeData[defaultOUIDKey]; ok && val != "" {
		ouRequest.Parent = &val
	}

	return ouRequest
}
