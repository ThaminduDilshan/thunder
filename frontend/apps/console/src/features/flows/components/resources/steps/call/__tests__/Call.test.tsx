/**
 * Copyright (c) 2026, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/* eslint-disable react/require-default-props, @typescript-eslint/no-unnecessary-type-assertion */

import {render, screen, fireEvent} from '@testing-library/react';
import {describe, it, expect, vi, beforeEach} from 'vitest';
import Call from '../Call';

// Mock i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (_key: string, fallback?: string) => fallback ?? _key,
  }),
}));

// Mock @xyflow/react
const mockDeleteElements = vi.fn();
const mockUseNodeId = vi.fn<() => string | null>(() => 'call-node-id');

vi.mock('@xyflow/react', () => ({
  Handle: ({type, position, id}: {type: string; position: string; id?: string}) => (
    <div data-testid={`handle-${type}-${id ?? position}`} data-position={position} data-handle-id={id} />
  ),
  Position: {
    Left: 'left',
    Right: 'right',
    Top: 'top',
    Bottom: 'bottom',
  },
  useNodeId: () => mockUseNodeId(),
  useReactFlow: () => ({
    deleteElements: mockDeleteElements,
  }),
}));

// Mock useInteractionState
const mockSetLastInteractedResource = vi.fn();
const mockSetLastInteractedStepId = vi.fn();
vi.mock('@/features/flows/hooks/useInteractionState', () => ({
  default: () => ({
    setLastInteractedResource: mockSetLastInteractedResource,
    setLastInteractedStepId: mockSetLastInteractedStepId,
  }),
}));

// Mock useUIPanelState
const mockSetIsOpenResourcePropertiesPanel = vi.fn();
vi.mock('@/features/flows/hooks/useUIPanelState', () => ({
  default: () => ({
    setIsOpenResourcePropertiesPanel: mockSetIsOpenResourcePropertiesPanel,
  }),
}));

describe('Call', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockUseNodeId.mockReturnValue('call-node-id');
  });

  describe('Rendering', () => {
    it('shows the configured flow ref', () => {
      render(<Call data={{flow: {ref: 'flow-123'}}} />);
      // `flow-123` appears in both the header (as part of "Call → flow-123") and the body.
      expect(screen.getAllByText(/flow-123/).length).toBeGreaterThan(0);
    });

    it('shows the "not configured" placeholder when no flow ref is set', () => {
      render(<Call data={{}} />);
      expect(screen.getByText(/not configured/i)).toBeInTheDocument();
    });

    it('shows the "Call flow" default label when no flow ref is set', () => {
      render(<Call data={{}} />);
      expect(screen.getByText('Call flow')).toBeInTheDocument();
    });
  });

  describe('Handles', () => {
    it('renders the target handle on the left', () => {
      render(<Call data={{flow: {ref: 'f1'}}} />);
      const target = screen.getByTestId('handle-target-left');
      expect(target).toHaveAttribute('data-position', 'left');
    });

    it('renders the success source handle on the right with the node-id suffix', () => {
      render(<Call data={{flow: {ref: 'f1'}}} />);
      // The handle id is `${stepId}${FLOW_BUILDER_NEXT_HANDLE_SUFFIX}` — we just need a right-positioned source handle.
      const handles = screen.getAllByTestId(/handle-source-/);
      const right = handles.find((h) => h.getAttribute('data-position') === 'right');
      expect(right).toBeTruthy();
    });

    it('renders the failure source handle on the bottom with id "failure"', () => {
      render(<Call data={{flow: {ref: 'f1'}}} />);
      const failure = screen.getByTestId('handle-source-failure');
      expect(failure).toHaveAttribute('data-position', 'bottom');
      expect(failure).toHaveAttribute('data-handle-id', 'failure');
    });
  });

  describe('Configure button', () => {
    it('opens the properties panel and sets the interacted resource on click', () => {
      const {container} = render(<Call data={{flow: {ref: 'flow-abc'}}} />);
      const configBtn = container.querySelectorAll('button')[0];
      expect(configBtn).toBeTruthy();
      fireEvent.click(configBtn!);
      expect(mockSetLastInteractedStepId).toHaveBeenCalledWith('call-node-id');
      expect(mockSetIsOpenResourcePropertiesPanel).toHaveBeenCalledWith(true);
      expect(mockSetLastInteractedResource).toHaveBeenCalled();
    });
  });

  describe('Delete button', () => {
    it('calls deleteElements with the node id', () => {
      const {container} = render(<Call data={{flow: {ref: 'flow-abc'}}} />);
      const deleteBtn = container.querySelectorAll('button')[1];
      expect(deleteBtn).toBeTruthy();
      fireEvent.click(deleteBtn!);
      expect(mockDeleteElements).toHaveBeenCalledWith({nodes: [{id: 'call-node-id'}]});
    });

    it('does not call deleteElements when nodeId is empty', () => {
      mockUseNodeId.mockReturnValue('');
      const {container} = render(<Call data={{}} />);
      const deleteBtn = container.querySelectorAll('button')[1];
      fireEvent.click(deleteBtn!);
      expect(mockDeleteElements).not.toHaveBeenCalled();
    });
  });
});
