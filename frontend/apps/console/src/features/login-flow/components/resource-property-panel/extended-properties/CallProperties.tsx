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

import {FormControl, FormLabel, MenuItem, Select, Stack, Typography} from '@wso2/oxygen-ui';
import {useMemo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import type {CommonResourcePropertiesPropsInterface} from './execution-properties/types';
import useGetFlows from '@/features/flows/api/useGetFlows';
import type {BasicFlowDefinition} from '@/features/flows/models/responses';
import type {StepData} from '@/features/flows/models/steps';

/**
 * Props interface of {@link CallProperties}
 */
export type CallPropertiesPropsInterface = CommonResourcePropertiesPropsInterface;

/**
 * Property panel for the CALL step type. Renders a flow-id dropdown sourced from the
 * `flows` API and writes the chosen id back to the step under `data.flow.ref`.
 * Branching to next nodes is configured by drawing canvas edges from the CALL node's
 * success (right) and failure (bottom) handles, so this panel only exposes the flow ref.
 */
function CallProperties({resource, onChange}: CallPropertiesPropsInterface): ReactElement {
  const {t} = useTranslation();
  const {data, isLoading, error} = useGetFlows({limit: 100});

  const currentRef = useMemo<string>(() => {
    const stepData = resource?.data as (StepData & {flow?: {ref?: string}}) | undefined;
    return stepData?.flow?.ref ?? '';
  }, [resource]);

  const flows: BasicFlowDefinition[] = useMemo<BasicFlowDefinition[]>(() => {
    const list = data?.flows ?? [];
    // Hide the current flow from the dropdown — a flow may not call itself.
    return list.filter((f: BasicFlowDefinition) => f.id !== resource?.id);
  }, [data, resource?.id]);

  const handleChange = (selected: string): void => {
    onChange('data.flow', {ref: selected}, resource);
  };

  if (error) {
    return (
      <Typography variant="body2" color="error" data-testid="call-properties-error">
        {t('flows:core.call.properties.loadError', 'Failed to load available flows')}
      </Typography>
    );
  }

  return (
    <Stack gap={2} data-testid="call-properties">
      <Typography variant="body2" color="text.secondary">
        {t('flows:core.call.properties.description', 'Pick the flow to invoke when this node executes.')}
      </Typography>
      <FormControl fullWidth size="small">
        <FormLabel htmlFor="call-flow-ref-select">
          {t('flows:core.call.properties.flow.label', 'Referenced flow')}
        </FormLabel>
        <Select
          id="call-flow-ref-select"
          data-testid="call-flow-ref-select"
          value={currentRef}
          disabled={isLoading || flows.length === 0}
          onChange={(e) => handleChange(String(e.target.value))}
          displayEmpty
          fullWidth
        >
          <MenuItem value="" disabled>
            {isLoading
              ? t('flows:core.call.properties.flow.loading', 'Loading flows…')
              : t('flows:core.call.properties.flow.placeholder', 'Select a flow')}
          </MenuItem>
          {flows.map((f: BasicFlowDefinition) => (
            <MenuItem key={f.id} value={f.id}>
              {f.name} ({f.flowType})
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Stack>
  );
}

export default CallProperties;
