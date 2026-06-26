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

/* eslint-disable @typescript-eslint/non-nullable-type-assertion-style */

import {fireEvent, render, screen} from '@testing-library/react';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import RichTextActionFields from '../RichTextActionFields';
import type {Resource} from '@/features/flows/models/resources';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (_key: string, fallback?: string) => fallback ?? _key,
  }),
}));

const makeResource = (overrides: Partial<Resource> = {}): Resource =>
  ({
    id: 'rt-1',
    type: 'RICH_TEXT',
    category: 'DISPLAY',
    resourceType: 'ELEMENT',
    ...overrides,
  }) as Resource;

describe('RichTextActionFields', () => {
  const onChange = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the action ref and event type fields', () => {
    render(<RichTextActionFields resource={makeResource()} onChange={onChange} />);
    expect(screen.getByText('Action ref')).toBeInTheDocument();
    expect(screen.getByText('Event type')).toBeInTheDocument();
  });

  it('shows the current action ref value when set', () => {
    const resource = makeResource({action: {ref: 'action_signup'}} as unknown as Partial<Resource>);
    render(<RichTextActionFields resource={resource} onChange={onChange} />);
    const input = screen.getByTestId('rich-text-action-ref').querySelector('input') as HTMLInputElement;
    expect(input.value).toBe('action_signup');
  });

  it('writes the new action ref via onChange with debounce', () => {
    const resource = makeResource();
    render(<RichTextActionFields resource={resource} onChange={onChange} />);
    const input = screen.getByTestId('rich-text-action-ref').querySelector('input') as HTMLInputElement;
    fireEvent.change(input, {target: {value: 'action_signup'}});
    expect(onChange).toHaveBeenCalledWith('action.ref', 'action_signup', resource, true);
  });

  it('disables the event-type select when no action ref is set', () => {
    render(<RichTextActionFields resource={makeResource()} onChange={onChange} />);
    const select = screen.getByTestId('rich-text-action-event-type');
    const button = select.querySelector('[role="combobox"]');
    expect(button).toHaveAttribute('aria-disabled', 'true');
  });

  it('enables the event-type select when an action ref is set', () => {
    const resource = makeResource({action: {ref: 'action_signup'}} as unknown as Partial<Resource>);
    render(<RichTextActionFields resource={resource} onChange={onChange} />);
    const select = screen.getByTestId('rich-text-action-event-type');
    const button = select.querySelector('[role="combobox"]');
    expect(button).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('writes the selected event type via onChange', () => {
    const resource = makeResource({action: {ref: 'action_signup'}} as unknown as Partial<Resource>);
    render(<RichTextActionFields resource={resource} onChange={onChange} />);
    const select = screen.getByTestId('rich-text-action-event-type');
    fireEvent.mouseDown(select.querySelector('[role="combobox"]')!);
    fireEvent.click(screen.getByText('SUBMIT'));
    expect(onChange).toHaveBeenCalledWith('action.eventType', 'SUBMIT', resource);
  });
});
