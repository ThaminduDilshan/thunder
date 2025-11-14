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

// Package flowmgt provides the flow management service implementation.
package flowmgt

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/asgardeo/thunder/internal/flow/common"
	"github.com/asgardeo/thunder/internal/flow/core"
	"github.com/asgardeo/thunder/internal/flow/executor"
	"github.com/asgardeo/thunder/internal/system/config"
	"github.com/asgardeo/thunder/internal/system/log"
	sysutils "github.com/asgardeo/thunder/internal/system/utils"
)

// FlowMgtServiceInterface defines the interface for the flow management service.
type FlowMgtServiceInterface interface {
	RegisterGraph(graphID string, g core.GraphInterface)
	GetGraph(graphID string) (core.GraphInterface, bool)
	IsValidGraphID(graphID string) bool
}

// flowMgtService is the implementation of FlowMgtServiceInterface.
type flowMgtService struct {
	graphs           map[string]core.GraphInterface
	mu               sync.Mutex
	flowFactory      core.FlowFactoryInterface
	executorRegistry executor.ExecutorRegistryInterface
	logger           *log.Logger
}

func newFlowMgtService(flowFactory core.FlowFactoryInterface,
	executorRegistry executor.ExecutorRegistryInterface) (FlowMgtServiceInterface, error) {
	flowMgtInstance := &flowMgtService{
		graphs:           make(map[string]core.GraphInterface),
		mu:               sync.Mutex{},
		flowFactory:      flowFactory,
		executorRegistry: executorRegistry,
		logger:           log.GetLogger().With(log.String(log.LoggerKeyComponentName, "FlowMgtService")),
	}
	err := flowMgtInstance.init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize flow management service: %w", err)
	}

	err = flowMgtInstance.validateDefaultFlowConfigs()
	if err != nil {
		return nil, errors.New("failed to validate default flow configurations: " + err.Error())
	}
	return flowMgtInstance, nil
}

// Init initializes the FlowMgtService by loading graph configurations into runtime.
func (s *flowMgtService) init() error {
	logger := s.logger
	logger.Debug("Initializing the flow management service")

	configDir := config.GetThunderRuntime().Config.Flow.GraphDirectory
	if configDir == "" {
		logger.Info("Graph directory is not set. No graphs will be loaded.")
		return nil
	}

	configDir = filepath.Join(config.GetThunderRuntime().ThunderHome, configDir)
	configDir = filepath.Clean(configDir)

	logger.Debug("Loading graphs from config directory", log.String("configDir", configDir))

	files, err := os.ReadDir(configDir)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Config directory does not exist. No graphs will be loaded.",
				log.String("configDir", configDir))
			return nil
		}
		return fmt.Errorf("failed to read config directory %s: %w", configDir, err)
	}

	if len(files) == 0 {
		logger.Info("No graph configuration files found in the configured directory. No graphs will be loaded.")
		return nil
	}
	logger.Debug("Found graph definition files in the graph directory", log.Int("fileCount", len(files)))

	// Process each JSON file in the directory
	flowGraphs := make(map[string]core.GraphInterface)
	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			logger.Debug("Skipping non-JSON file or directory",
				log.String("fileName", file.Name()), log.Bool("isDir", file.IsDir()))
			continue
		}
		filePath := filepath.Join(configDir, file.Name())
		filePath = filepath.Clean(filePath)

		// Read the file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			logger.Warn("Failed to read graph file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Parse the JSON into the flow model
		var jsonGraph graphDefinition
		if err := json.Unmarshal(fileContent, &jsonGraph); err != nil {
			logger.Warn("Failed to parse JSON in file", log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Convert the JSON graph definition to the graph model
		graphModel, err := s.buildGraphFromDefinition(&jsonGraph)
		if err != nil {
			logger.Warn("Failed to convert graph definition to graph model",
				log.String("filePath", filePath), log.Error(err))
			continue
		}

		// Log the graph model as JSON for debugging
		if logger.IsDebugEnabled() {
			jsonString, err := graphModel.ToJSON()
			if err != nil {
				logger.Warn("Failed to convert graph model to JSON", log.String("filePath", filePath), log.Error(err))
			} else {
				logger.Debug("Graph model loaded successfully", log.String("graphID", graphModel.GetID()),
					log.String("json", jsonString))
			}
		}

		// Append graph to the flowGraphs map
		flowGraphs[graphModel.GetID()] = graphModel
	}

	// Register all loaded graphs
	inferredGraphCount := 0
	for graphID, graph := range flowGraphs {
		// Create and register the equivalent registration graph if not found already.
		registrationGraphID := s.getRegistrationGraphID(graphID)
		_, exists := s.graphs[registrationGraphID]
		if !exists && graph.GetType() == common.FlowTypeAuthentication {
			if err := s.createAndRegisterRegistrationGraph(registrationGraphID, graph); err != nil {
				logger.Error("Failed creating registration graph", log.String("graphID", graphID), log.Error(err))
				continue
			}
			inferredGraphCount++
		}

		logger.Debug("Registering graph", log.String("graphType", string(graph.GetType())),
			log.String("graphID", graphID))
		s.RegisterGraph(graphID, graph)
	}

	logger.Debug("Flow management service initialized successfully", log.Int("configuredGraphCount", len(flowGraphs)),
		log.Int("inferredGraphCount", inferredGraphCount))

	return nil
}

// RegisterGraph registers a graph with the FlowMgtService by its ID.
func (s *flowMgtService) RegisterGraph(graphID string, g core.GraphInterface) {
	s.graphs[graphID] = g
}

// GetGraph retrieves a graph by its ID
func (s *flowMgtService) GetGraph(graphID string) (core.GraphInterface, bool) {
	g, ok := s.graphs[graphID]
	return g, ok
}

// IsValidGraphID checks if the provided graph ID is valid and exists in the service.
func (s *flowMgtService) IsValidGraphID(graphID string) bool {
	if graphID == "" {
		return false
	}
	_, exists := s.graphs[graphID]
	return exists
}

// buildGraphFromDefinition builds a graph from a graph definition json.
func (s *flowMgtService) buildGraphFromDefinition(definition *graphDefinition) (core.GraphInterface, error) {
	if definition == nil || len(definition.Nodes) == 0 {
		return nil, fmt.Errorf("graph definition is nil or has no nodes")
	}

	// Create a graph
	_type, err := getGraphType(definition.Type)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving graph type: %w", err)
	}
	g := s.flowFactory.CreateGraph(definition.ID, _type)

	// Add all nodes to the graph
	edges := make(map[string][]string)
	for _, nodeDef := range definition.Nodes {
		isFinalNode := len(nodeDef.Next) == 0

		// Construct a new node. Here we set isStartNode to false by default.
		node, err := s.flowFactory.CreateNode(nodeDef.ID, nodeDef.Type, nodeDef.Properties,
			false, isFinalNode)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeDef.ID, err)
		}

		// Set next nodes if defined
		if len(nodeDef.Next) > 0 {
			node.SetNextNodeList(nodeDef.Next)

			// Store edges based on the node definition
			_, exists := edges[nodeDef.ID]
			if !exists {
				edges[nodeDef.ID] = []string{}
			}
			edges[nodeDef.ID] = append(edges[nodeDef.ID], nodeDef.Next...)
		}

		// Convert and set input data from definition
		inputData := make([]common.InputData, len(nodeDef.InputData))
		for i, input := range nodeDef.InputData {
			inputData[i] = common.InputData{
				Name:     input.Name,
				Type:     input.Type,
				Required: input.Required,
			}
		}
		node.SetInputData(inputData)

		// Set the executor name if defined, otherwise check for special node types
		executorName := nodeDef.Executor.Name
		if executorName == "" {
			// Check if it is the auth assert node
			if nodeDef.Type == string(common.NodeTypeAuthSuccess) {
				executorName = executor.ExecutorNameAuthAssert
			} else if nodeDef.Type == string(common.NodeTypeRegistrationStart) {
				executorName = executor.ExecutorNameUserTypeResolver
			}
		}

		// Set node executor if defined
		if executorName != "" {
			err := s.validateExecutorName(executorName)
			if err != nil {
				return nil, fmt.Errorf("error while validating executor %s: %w", executorName, err)
			}
			if executableNode, ok := node.(core.ExecutorBackedNodeInterface); ok {
				executableNode.SetExecutorName(executorName)
			} else {
				return nil, fmt.Errorf("node %s of type %s does not support executors", nodeDef.ID, nodeDef.Type)
			}
		}

		err = g.AddNode(node)
		if err != nil {
			return nil, fmt.Errorf("failed to add node %s to the graph: %w", nodeDef.ID, err)
		}
	}

	// Set edges in the graph
	for sourceID, targetIDs := range edges {
		for _, targetID := range targetIDs {
			err := g.AddEdge(sourceID, targetID)
			if err != nil {
				return nil, fmt.Errorf("failed to add edge from %s to %s: %w", sourceID, targetID, err)
			}
		}
	}

	// Determine the start node and set it in the graph
	startNodeID := ""
	for _, node := range g.GetNodes() {
		if len(node.GetPreviousNodeList()) == 0 {
			startNodeID = node.GetID()
			break
		}
	}
	if startNodeID == "" {
		return nil, fmt.Errorf("no start node found in the graph definition")
	}

	err = g.SetStartNode(startNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to set start node ID: %w", err)
	}

	return g, nil
}

// validateExecutorName validates that an executor with the given name is registered.
func (s *flowMgtService) validateExecutorName(executorName string) error {
	if executorName == "" {
		return fmt.Errorf("executor name cannot be empty")
	}
	if !s.executorRegistry.IsRegistered(executorName) {
		return fmt.Errorf("executor with name %s not registered", executorName)
	}

	return nil
}

// getRegistrationGraphID constructs the registration graph ID from the auth graph ID.
func (s *flowMgtService) getRegistrationGraphID(authGraphID string) string {
	return common.RegistrationFlowGraphPrefix + strings.TrimPrefix(authGraphID, common.AuthFlowGraphPrefix)
}

// createAndRegisterRegistrationGraph creates a registration graph from an authentication graph and registers it.
func (s *flowMgtService) createAndRegisterRegistrationGraph(
	registrationGraphID string, authGraph core.GraphInterface) error {
	logger := s.logger
	registrationGraph, err := s.createRegistrationGraph(registrationGraphID, authGraph)
	if err != nil {
		return fmt.Errorf("failed to infer registration graph: %w", err)
	}

	if logger.IsDebugEnabled() {
		registrationGraphJSON, err := registrationGraph.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON", log.String("graphID", registrationGraphID),
				log.Error(err))
		} else {
			logger.Debug("Graph model loaded successfully", log.String("graphID", registrationGraph.GetID()),
				log.String("json", registrationGraphJSON))
		}
	}

	logger.Debug("Registering inferred registration graph", log.String("graphID", registrationGraph.GetID()))
	s.RegisterGraph(registrationGraph.GetID(), registrationGraph)
	return nil
}

// createRegistrationGraph creates a registration graph from an authentication graph.
func (s *flowMgtService) createRegistrationGraph(registrationGraphID string,
	authGraph core.GraphInterface) (core.GraphInterface, error) {
	logger := s.logger
	logger.Debug("Creating registration graph from authentication graph",
		log.String("registrationGraphID", registrationGraphID),
		log.String("authGraphID", authGraph.GetID()))

	// Create a new graph from the authentication graph
	registrationGraph := s.flowFactory.CreateGraph(registrationGraphID, common.FlowTypeRegistration)
	nodesCopy, err := s.flowFactory.CloneNodes(authGraph.GetNodes())
	if err != nil {
		return nil, fmt.Errorf("failed to clone nodes from auth graph: %w", err)
	}
	registrationGraph.SetNodes(nodesCopy)
	registrationGraph.SetEdges(sysutils.DeepCopyMapOfStringSlices(authGraph.GetEdges()))

	// Find authentication success node to insert provisioning node before it
	authSuccessNodeID := ""
	nodes := registrationGraph.GetNodes()
	for nodeID, node := range nodes {
		if node.IsFinalNode() {
			authSuccessNodeID = nodeID
			break
		}
	}
	if authSuccessNodeID == "" {
		return nil, fmt.Errorf("no authentication success node found in the authentication graph")
	}

	// Create and add provisioning node
	provisioningNode, err := s.createProvisioningNode()
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}
	err = registrationGraph.AddNode(provisioningNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add provisioning node to registration graph: %w", err)
	}

	// Modify the edges that lead to the auth success node to point to the provisioning node
	for fromNodeID, toNodeIDs := range registrationGraph.GetEdges() {
		for _, toNodeID := range toNodeIDs {
			if toNodeID == authSuccessNodeID {
				err := registrationGraph.RemoveEdge(fromNodeID, toNodeID)
				if err != nil {
					return nil, fmt.Errorf("failed to remove edge from %s to %s: %w", fromNodeID, toNodeID, err)
				}

				err = registrationGraph.AddEdge(fromNodeID, provisioningNode.GetID())
				if err != nil {
					return nil, fmt.Errorf("failed to add edge from %s to provisioning node: %w", fromNodeID, err)
				}
			}
		}
	}

	// Add an edge from the provisioning node to the auth success node
	err = registrationGraph.AddEdge(provisioningNode.GetID(), authSuccessNodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to add edge from provisioning node to auth success node: %w", err)
	}

	// Insert the registration start node
	regStartNode, err := s.createRegistrationStartNode()
	if err != nil {
		return nil, fmt.Errorf("failed to create registration start node: %w", err)
	}
	err = registrationGraph.AddNode(regStartNode)
	if err != nil {
		return nil, fmt.Errorf("failed to add registration start node: %w", err)
	}

	// Add edge from registration start to the previous start node
	if err := registrationGraph.AddEdge(regStartNode.GetID(), authGraph.GetStartNodeID()); err != nil {
		return nil, fmt.Errorf("failed to connect registration start to previous start: %w", err)
	}
	if err := registrationGraph.SetStartNode(regStartNode.GetID()); err != nil {
		return nil, fmt.Errorf("failed to set registration start node as start: %w", err)
	}

	// Log the graph model as JSON for debugging
	if logger.IsDebugEnabled() {
		jsonString, err := registrationGraph.ToJSON()
		if err != nil {
			logger.Warn("Failed to convert graph model to JSON",
				log.String("graphID", registrationGraph.GetID()), log.Error(err))
		} else {
			logger.Debug("Registration graph model loaded successfully",
				log.String("graphID", registrationGraph.GetID()), log.String("json", jsonString))
		}
	}

	return registrationGraph, nil
}

// createProvisioningNode creates a provisioning node that leads to the specified auth success node
func (s *flowMgtService) createProvisioningNode() (core.NodeInterface, error) {
	provisioningNode, err := s.flowFactory.CreateNode(
		"provisioning",
		string(common.NodeTypeTaskExecution),
		map[string]interface{}{},
		false,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create provisioning node: %w", err)
	}

	// Set executor name for the provisioning node
	if executableNode, ok := provisioningNode.(core.ExecutorBackedNodeInterface); ok {
		executableNode.SetExecutorName(executor.ExecutorNameProvisioning)
	} else {
		return nil, fmt.Errorf("provisioning node does not implement ExecutorBackedNodeInterface")
	}

	return provisioningNode, nil
}

// createRegistrationStartNode creates the registration start node.
func (s *flowMgtService) createRegistrationStartNode() (core.NodeInterface, error) {
	regStartNode, err := s.flowFactory.CreateNode(
		"registration_start",
		string(common.NodeTypeRegistrationStart),
		map[string]interface{}{},
		false,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create registration start node: %w", err)
	}

	// Set executor name for the registration start node
	if executableNode, ok := regStartNode.(core.ExecutorBackedNodeInterface); ok {
		executableNode.SetExecutorName(executor.ExecutorNameUserTypeResolver)
	} else {
		return nil, fmt.Errorf("registration start node does not implement ExecutorBackedNodeInterface")
	}

	return regStartNode, nil
}

// validateDefaultFlowConfigs validates the default flow configurations.
func (s *flowMgtService) validateDefaultFlowConfigs() error {
	flowConfig := config.GetThunderRuntime().Config.Flow

	// Validate auth flow.
	if flowConfig.Authn.DefaultFlow == "" {
		return errors.New("default authentication flow is not configured")
	}
	if !s.IsValidGraphID(flowConfig.Authn.DefaultFlow) {
		return errors.New("default authentication flow graph ID is invalid")
	}

	return nil
}
