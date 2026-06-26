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

import {describe, expect, it} from 'vitest';
import {
  EmbeddedFlowComponent,
  EmbeddedFlowComponentAction,
  EmbeddedFlowComponentType,
  EmbeddedFlowEventType,
} from '../embedded-flow-v2';

describe('embedded-flow-v2 RICH_TEXT action wiring', () => {
  it('accepts a RICH_TEXT component without an action (pure display, current behavior)', () => {
    const component: EmbeddedFlowComponent = {
      id: 'rt_display_only',
      label: '<p>Pure display rich text</p>',
      type: EmbeddedFlowComponentType.RichText,
    };
    expect(component.action).toBeUndefined();
    expect(component.type).toBe(EmbeddedFlowComponentType.RichText);
  });

  it('accepts an action with a ref only (no eventType)', () => {
    const action: EmbeddedFlowComponentAction = {ref: 'action_signup'};
    expect(action.ref).toBe('action_signup');
    expect(action.eventType).toBeUndefined();
  });

  it('accepts an action with a ref and an eventType', () => {
    const action: EmbeddedFlowComponentAction = {
      eventType: EmbeddedFlowEventType.Submit,
      ref: 'action_signup',
    };
    expect(action.ref).toBe('action_signup');
    expect(action.eventType).toBe(EmbeddedFlowEventType.Submit);
  });

  it('accepts a RICH_TEXT component carrying an action (interactive variant)', () => {
    const component: EmbeddedFlowComponent = {
      action: {
        eventType: EmbeddedFlowEventType.Submit,
        ref: 'action_signup',
      },
      id: 'rt_signup',
      label: '<p>Don\'t have an account? <a data-action-ref="action_signup">Sign up</a></p>',
      type: EmbeddedFlowComponentType.RichText,
    };
    expect(component.action).toBeDefined();
    expect(component.action?.ref).toBe('action_signup');
    expect(component.action?.eventType).toBe(EmbeddedFlowEventType.Submit);
  });
});
