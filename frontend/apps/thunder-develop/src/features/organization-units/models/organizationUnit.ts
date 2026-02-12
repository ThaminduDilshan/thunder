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

import type {Dispatch, SetStateAction} from 'react';

export interface OUTreeItem {
  id: string;
  label: string;
  handle: string;
  description?: string | null;
  isPlaceholder?: boolean;
  children?: OUTreeItem[];
}

export interface OrganizationUnitContextType {
  treeItems: OUTreeItem[];
  setTreeItems: Dispatch<SetStateAction<OUTreeItem[]>>;
  expandedItems: string[];
  setExpandedItems: Dispatch<SetStateAction<string[]>>;
  loadedItems: Set<string>;
  setLoadedItems: Dispatch<SetStateAction<Set<string>>>;
  /** Clears treeItems and loadedItems, forcing a re-fetch. Preserves expandedItems so the tree re-expands after rebuild. */
  resetTreeState: () => void;
}
