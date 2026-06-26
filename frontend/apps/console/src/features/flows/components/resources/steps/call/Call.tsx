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

import {Box, Card, IconButton, Tooltip, Typography} from '@wso2/oxygen-ui';
import {CogIcon, TrashIcon} from '@wso2/oxygen-ui-icons-react';
import {Handle, Position, useNodeId, useReactFlow} from '@xyflow/react';
import {memo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import VisualFlowConstants from '@/features/flows/constants/VisualFlowConstants';
import useInteractionState from '@/features/flows/hooks/useInteractionState';
import useUIPanelState from '@/features/flows/hooks/useUIPanelState';
import {ResourceTypes} from '@/features/flows/models/resources';
import {type Step, type StepData, StepCategories, StepTypes} from '@/features/flows/models/steps';

/**
 * Props interface of {@link Call}
 */
export interface CallPropsInterface {
  data?: StepData & {flow?: {ref?: string}};
  resources?: Step[];
}

/**
 * Call node component for cross-flow invocation. Visually mirrors ExecutionMinimal but
 * exposes a flow reference instead of an executor and exposes both `onSuccess` (right) and
 * `onFailure` (bottom) handles.
 */
function Call({data = undefined, resources = []}: CallPropsInterface): ReactElement {
  const stepId: string | null = useNodeId();
  const {t} = useTranslation();
  const {setLastInteractedResource, setLastInteractedStepId} = useInteractionState();
  const {setIsOpenResourcePropertiesPanel} = useUIPanelState();
  const {deleteElements} = useReactFlow();

  const flowRef: string = data?.flow?.ref ?? '';
  const displayLabel: string =
    resources[0]?.display?.label ?? (flowRef ? `Call → ${flowRef}` : t('flows:core.call.unconfiguredLabel', 'Call flow'));

  const resource: Step = {
    id: stepId ?? '',
    type: StepTypes.Call,
    category: StepCategories.Workflow,
    resourceType: ResourceTypes.Step,
    data,
    display: {
      label: displayLabel,
      showOnResourcePanel: false,
    },
  } as Step;

  const handleConfigClick = (): void => {
    if (stepId !== null) {
      setLastInteractedStepId(stepId);
    }
    setLastInteractedResource(resource);
    setIsOpenResourcePropertiesPanel(true);
  };

  return (
    <Box className="call-step has-branching" data-testid="call-node">
      <Box
        display="flex"
        justifyContent="space-between"
        alignItems="center"
        className="call-step-action-panel"
        sx={{
          backgroundColor: 'secondary.main',
          px: 2,
          py: 1.25,
          height: 44,
        }}
      >
        <Typography variant="body2" sx={{color: 'common.white', fontWeight: 500}}>
          {displayLabel}
        </Typography>
        <Box display="flex" alignItems="center" gap={0.5}>
          <Tooltip title={t('flows:core.call.tooltip.configure', 'Configure')}>
            <IconButton size="small" onClick={handleConfigClick} sx={{color: 'common.white'}}>
              <CogIcon size={18} />
            </IconButton>
          </Tooltip>
          <Tooltip title={t('flows:core.call.tooltip.delete', 'Delete')}>
            <IconButton
              size="small"
              onClick={() => {
                if (stepId) {
                  // eslint-disable-next-line @typescript-eslint/no-floating-promises
                  deleteElements({nodes: [{id: stepId}]});
                }
              }}
              sx={{color: 'common.white'}}
            >
              <TrashIcon size={18} />
            </IconButton>
          </Tooltip>
        </Box>
      </Box>
      <Handle type="target" position={Position.Left} />
      <Card
        className="call-step-content"
        onClick={() => {
          if (stepId) {
            setLastInteractedStepId(stepId);
          }
          setLastInteractedResource(resource);
        }}
        sx={{p: 2}}
      >
        <Typography variant="caption" sx={{display: 'block', opacity: 0.7}}>
          {t('flows:core.call.referencedFlow', 'Referenced flow')}
        </Typography>
        <Typography variant="body2" sx={{wordBreak: 'break-all'}}>
          {flowRef || t('flows:core.call.notConfigured', '— not configured —')}
        </Typography>
      </Card>
      <Tooltip title={t('flows:core.call.handles.success', 'On success')} placement="right">
        <Box className="handle-wrapper success-wrapper">
          <Handle
            type="source"
            position={Position.Right}
            id={`${stepId ?? ''}${VisualFlowConstants.FLOW_BUILDER_NEXT_HANDLE_SUFFIX}`}
            className="call-handle-success"
          />
        </Box>
      </Tooltip>
      <Tooltip title={t('flows:core.call.handles.failure', 'On failure')} placement="bottom">
        <Box className="handle-wrapper failure-wrapper">
          <Handle type="source" position={Position.Bottom} id="failure" className="call-handle-failure" />
        </Box>
      </Tooltip>
    </Box>
  );
}

export default memo(Call);
