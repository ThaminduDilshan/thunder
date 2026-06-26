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

import {FormControl, FormLabel, MenuItem, Select, Stack, TextField, Typography} from '@wso2/oxygen-ui';
import {useMemo, type ReactElement} from 'react';
import {useTranslation} from 'react-i18next';
import type {Resource} from '../../../models/resources';

const EVENT_TYPE_VALUES = ['SUBMIT', 'TRIGGER'] as const;

/**
 * Action wiring for a RICH_TEXT element. Mirrors the SDK `EmbeddedFlowComponentAction`
 * shape: `{ref, eventType?}`.
 */
export interface RichTextAction {
  ref?: string;
  eventType?: string;
}

/**
 * Props interface of {@link RichTextActionFields}
 */
export interface RichTextActionFieldsProps {
  /** The RICH_TEXT element whose `action` field is being authored. */
  resource: Resource;
  /** Callback fired when either the action ref or event type changes. */
  onChange: (propertyKey: string, newValue: unknown, resource: Resource, debounce?: boolean) => void;
}

/**
 * Property-panel UI for the optional `action` field on a RICH_TEXT element. Lets the
 * author configure the action ref (which must match a `data-action-ref="<ref>"` attribute
 * on an anchor inside the rich-text HTML and a prompt-level action ref) and the event
 * type dispatched by the SDK renderer.
 */
function RichTextActionFields({resource, onChange}: RichTextActionFieldsProps): ReactElement {
  const {t} = useTranslation();

  const action: RichTextAction = useMemo<RichTextAction>(() => {
    const r = resource as Resource & {action?: RichTextAction};
    return r.action ?? {};
  }, [resource]);

  const ref: string = action.ref ?? '';
  const eventType: string = action.eventType ?? '';

  return (
    <Stack gap={2} data-testid="rich-text-action-fields">
      <Typography variant="body2" color="text.secondary">
        {t(
          'flows:core.elements.richText.action.description',
          'Optional. Wire a sentinel-marked anchor (data-action-ref) to a flow action.',
        )}
      </Typography>
      <FormControl fullWidth size="small">
        <FormLabel htmlFor="rich-text-action-ref">
          {t('flows:core.elements.richText.action.ref.label', 'Action ref')}
        </FormLabel>
        <TextField
          id="rich-text-action-ref"
          data-testid="rich-text-action-ref"
          value={ref}
          placeholder={t('flows:core.elements.richText.action.ref.placeholder', 'e.g. action_signup')}
          onChange={(e) => onChange('action.ref', e.target.value, resource, true)}
          size="small"
          fullWidth
        />
      </FormControl>
      <FormControl fullWidth size="small">
        <FormLabel htmlFor="rich-text-action-event-type">
          {t('flows:core.elements.richText.action.eventType.label', 'Event type')}
        </FormLabel>
        <Select
          id="rich-text-action-event-type"
          data-testid="rich-text-action-event-type"
          value={eventType}
          displayEmpty
          disabled={!ref}
          onChange={(e) => onChange('action.eventType', String(e.target.value), resource)}
        >
          <MenuItem value="">
            {t('flows:core.elements.richText.action.eventType.none', '— None —')}
          </MenuItem>
          {EVENT_TYPE_VALUES.map((value) => (
            <MenuItem key={value} value={value}>
              {value}
            </MenuItem>
          ))}
        </Select>
      </FormControl>
    </Stack>
  );
}

export default RichTextActionFields;
