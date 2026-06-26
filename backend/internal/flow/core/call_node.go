/*
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
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

package core

import (
	"github.com/thunder-id/thunderid/internal/flow/common"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
	"github.com/thunder-id/thunderid/internal/system/log"
)

// CallNodeInterface extends NodeInterface for CALL nodes, which transfer execution to
// a referenced flow and return control to the caller when the callee's END node is reached.
type CallNodeInterface interface {
	NodeInterface
	GetReferencedFlow() string
	SetReferencedFlow(flowID string)
	GetOnSuccess() string
	SetOnSuccess(nodeID string)
	GetOnFailure() string
	SetOnFailure(nodeID string)
}

// callNode implements CallNodeInterface.
type callNode struct {
	*node
	referencedFlow string
	onSuccess      string
	onFailure      string
	logger         *log.Logger
}

// Ensure callNode implements CallNodeInterface.
var _ CallNodeInterface = (*callNode)(nil)

// newCallNode creates a new CALL node.
func newCallNode(id string, properties map[string]interface{}, isStartNode, isFinalNode bool) NodeInterface {
	if properties == nil {
		properties = make(map[string]interface{})
	}
	return &callNode{
		node: &node{
			id:               id,
			_type:            common.NodeTypeCall,
			properties:       properties,
			isStartNode:      isStartNode,
			isFinalNode:      isFinalNode,
			nextNodeList:     []string{},
			previousNodeList: []string{},
		},
		logger: log.GetLogger().With(log.String(log.LoggerKeyComponentName, "CallNode"),
			log.String(log.LoggerKeyNodeID, id)),
	}
}

// Execute signals the engine to push a frame and transfer execution to the referenced flow.
func (n *callNode) Execute(ctx *NodeContext) (*common.NodeResponse, *serviceerror.ServiceError) {
	if n.referencedFlow == "" {
		n.logger.Error(ctx.Context, "Referenced flow ID is not set for CALL node")
		return nil, &serviceerror.InternalServerError
	}

	return &common.NodeResponse{
		Status:           common.NodeStatusCall,
		CallTargetFlowID: n.referencedFlow,
	}, nil
}

// GetReferencedFlow returns the ID of the flow this node calls.
func (n *callNode) GetReferencedFlow() string {
	return n.referencedFlow
}

// SetReferencedFlow sets the ID of the flow this node calls.
func (n *callNode) SetReferencedFlow(flowID string) {
	n.referencedFlow = flowID
}

// GetOnSuccess returns the caller node ID to proceed to when the callee flow ends successfully.
func (n *callNode) GetOnSuccess() string {
	return n.onSuccess
}

// SetOnSuccess sets the caller node ID to proceed to when the callee flow ends successfully.
func (n *callNode) SetOnSuccess(nodeID string) {
	n.onSuccess = nodeID
}

// GetOnFailure returns the caller node ID to proceed to when the callee flow ends with failure.
func (n *callNode) GetOnFailure() string {
	return n.onFailure
}

// SetOnFailure sets the caller node ID to proceed to when the callee flow ends with failure.
func (n *callNode) SetOnFailure(nodeID string) {
	n.onFailure = nodeID
}
