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
	"github.com/asgardeo/thunder/internal/flow/constants"
	"github.com/asgardeo/thunder/internal/system/error/serviceerror"
)

type DecisionNode struct {
	*Node
}

func newDecisionNode(id string, isStartNode bool, isFinalNode bool) NodeInterface {
	return &DecisionNode{
		Node: &Node{
			id:             id,
			_type:          constants.NodeTypeDecision,
			isStartNode:    isStartNode,
			isFinalNode:    isFinalNode,
			nextNodeID:     "",
			previousNodeID: "",
			inputData:      []InputData{},
			executorConfig: nil,
		},
	}
}

func (n *DecisionNode) Execute(ctx *NodeContext) (*NodeResponse, *serviceerror.ServiceError) {
	triggeredActionID := ctx.CurrentActionID
	if triggeredActionID != "" {
		// Loop through edges and find the selected node ID based on the action ID.
		//     - If found set next node ID and return.
		//     - Set current action id to nil when returning.
	}

	// If no next node found, return success completion.

	// Else return Incomplete with type VIEW.

	// TODO: Responsibility is to find next nodes via edges and return the next node IDs.
	return nil, nil
}
