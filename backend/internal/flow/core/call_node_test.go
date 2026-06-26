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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/thunder-id/thunderid/internal/flow/common"
	"github.com/thunder-id/thunderid/internal/system/error/serviceerror"
)

type CallNodeTestSuite struct {
	suite.Suite
}

func TestCallNodeTestSuite(t *testing.T) {
	suite.Run(t, new(CallNodeTestSuite))
}

func (s *CallNodeTestSuite) TestNewCallNode_ImplementsCallNodeInterface() {
	node := newCallNode("call-1", nil, false, false)
	_, ok := node.(CallNodeInterface)
	s.True(ok, "newCallNode should return a CallNodeInterface")
}

func (s *CallNodeTestSuite) TestNewCallNode_Defaults() {
	node := newCallNode("call-1", nil, false, false)
	s.Equal("call-1", node.GetID())
	s.Equal(common.NodeTypeCall, node.GetType())
	s.False(node.IsStartNode())
	s.False(node.IsFinalNode())
	s.NotNil(node.GetProperties())
}

func (s *CallNodeTestSuite) TestNewCallNode_WithProperties() {
	props := map[string]interface{}{"key": "value"}
	node := newCallNode("call-2", props, true, false)
	s.Equal("call-2", node.GetID())
	s.True(node.IsStartNode())
	s.Equal(props, node.GetProperties())
}

func (s *CallNodeTestSuite) TestCallNode_GetSetReferencedFlow() {
	node := newCallNode("call-1", nil, false, false)
	cn := node.(CallNodeInterface)

	s.Equal("", cn.GetReferencedFlow())
	cn.SetReferencedFlow("flow-abc")
	s.Equal("flow-abc", cn.GetReferencedFlow())
}

func (s *CallNodeTestSuite) TestCallNode_GetSetOnSuccess() {
	node := newCallNode("call-1", nil, false, false)
	cn := node.(CallNodeInterface)

	s.Equal("", cn.GetOnSuccess())
	cn.SetOnSuccess("next-node")
	s.Equal("next-node", cn.GetOnSuccess())
}

func (s *CallNodeTestSuite) TestCallNode_GetSetOnFailure() {
	node := newCallNode("call-1", nil, false, false)
	cn := node.(CallNodeInterface)

	s.Equal("", cn.GetOnFailure())
	cn.SetOnFailure("error-node")
	s.Equal("error-node", cn.GetOnFailure())
}

func (s *CallNodeTestSuite) TestCallNode_Execute_ReturnsCallStatus() {
	node := newCallNode("call-1", nil, false, false)
	cn := node.(CallNodeInterface)
	cn.SetReferencedFlow("target-flow-id")

	resp, svcErr := node.Execute(nil)

	s.Nil(svcErr)
	s.NotNil(resp)
	s.Equal(common.NodeStatusCall, resp.Status)
	s.Equal("target-flow-id", resp.CallTargetFlowID)
}

func (s *CallNodeTestSuite) TestCallNode_Execute_EmptyRef_ReturnsError() {
	node := newCallNode("call-1", nil, false, false)

	resp, svcErr := node.Execute(&NodeContext{Context: context.Background()})

	s.Nil(resp)
	s.NotNil(svcErr)
	s.Equal(serviceerror.InternalServerError.Code, svcErr.Code)
}
