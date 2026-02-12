/**
 * Copyright (c) 2025-2026, WSO2 LLC. (https://www.wso2.com).
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

import type {JSX} from 'react';
import {Box, Stack, Typography} from '@wso2/oxygen-ui';
import {useTranslation} from 'react-i18next';
import OrganizationUnitsTreeView from '../components/OrganizationUnitsTreeView';

export default function OrganizationUnitsListPage(): JSX.Element {
  const {t} = useTranslation();

  return (
    <Box>
      <Stack direction="row" justifyContent="space-between" alignItems="center" mb={4} flexWrap="wrap" gap={2}>
        <Box>
          <Typography variant="h1" gutterBottom>
            {t('organizationUnits:listing.title')}
          </Typography>
          <Typography variant="subtitle1" color="text.secondary">
            {t('organizationUnits:listing.subtitle')}
          </Typography>
        </Box>
      </Stack>

      <OrganizationUnitsTreeView />
    </Box>
  );
}
