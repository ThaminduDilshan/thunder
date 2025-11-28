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

import { useEffect, useState, useMemo, useRef } from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import { exchangeCodeForToken, initiateAuthWithPKCE, getAndClearCodeVerifier, validateState } from '../services/authService';
import useAuth from '../hooks/useAuth';
import ErrorPage from './ErrorPage';
import config from '../config';
import Layout from '../components/Layout';
import GradientCircularProgress from '../components/GradientCircularProgress';

interface TokenErrorInterface {
    error: string;
    error_description: string;
}

const RedirectLoginPage = () => {

    const { token, setToken } = useAuth();
    const [ error, setError ] = useState<TokenErrorInterface | null>(null);
    const [ loading, setLoading ] = useState<boolean>(false);

    const hasFetched = useRef(false);

    const urlParams = useMemo(() => new URLSearchParams(window.location.search), []);

    const handleLogin = async () => {
        setLoading(true);
        try {
            const { clientId, redirectUri, authorizationEndpoint, scope } = config;
            await initiateAuthWithPKCE(authorizationEndpoint, clientId, redirectUri, scope);
        } catch (error) {
            console.error('Error initiating OAuth flow:', error);
            setError({
                error: 'initialization_error',
                error_description: 'Failed to initiate OAuth flow. Please try again.'
            });
        } finally {
            // Reset loading state if redirect failed
            setLoading(false);
        }
    };

    useEffect(() => {
        const error = urlParams.get('error');
        const errorDescription = urlParams.get('error_description');

        if (error) {
        setError({
            error: error,
            error_description: errorDescription || 'No description available',
        });

        return;
        }
    }, [urlParams]);

    useEffect(() => {
        // Prevent double fetch calls
        if (hasFetched.current) return;
        hasFetched.current = true;

        const code = urlParams.get('code');
        const state = urlParams.get('state');
        
        if (code && !token) {
            // Validate state parameter for CSRF protection
            if (state && !validateState(state)) {
                console.error('State validation failed - possible CSRF attack');
                setError({
                    error: 'invalid_state',
                    error_description: 'State validation failed. Please try again.'
                });
                return;
            }
            
            setLoading(true);
            // Retrieve the code verifier from session storage for PKCE
            const codeVerifier = getAndClearCodeVerifier();
            
            exchangeCodeForToken(code, codeVerifier)
                .then(response => {
                    setToken(response.access_token);
                    // Clean up URL by removing query parameters after successful token exchange
                    window.history.replaceState({}, document.title, window.location.pathname);
                })
                .catch(error => {
                    console.error('Error fetching access token:', error);
                    if (error.response && error.response.data) {
                        setError(error.response.data);
                    } else {
                        setError({
                            error: 'unknown_error',
                            error_description: 'An unexpected error occurred during authentication. Please try again.'
                        });
                    }
                })
                .finally(() => {
                    setLoading(false);
                });
        }
    }, [token, urlParams, setToken]);

    if (error) {
        return (
            <ErrorPage
                errorCode={error.error || 'Unknown Error'}
                errorMessage={error.error_description || 'No description available'}
            />
        );
    }

    return (
        <Layout>
            {loading ? (
                <GradientCircularProgress />
            ) : (
                <Grid size={{ xs: 12, md: 6 }}>
                    <Paper
                        sx={{
                            display: "flex",
                            width: "100%",
                            height: "100%",
                            flexDirection: "column",
                        }}
                    >
                        <Box
                            sx={{
                                px: 5,
                                pt: 8,
                                pb: 5,
                                display: "flex",
                                flexDirection: "column",
                                alignItems: "flex-start",
                                gap: 2,
                            }}
                        >
                            <Typography variant="h4" component="h1" gutterBottom>
                                OAuth Sample App
                            </Typography>
                            <Typography variant="body1" color="text.secondary" sx={{ mb: 2 }}>
                                Sign in to your account to continue
                            </Typography>
                            
                            <Button
                                fullWidth
                                variant="contained"
                                color="primary"
                                size="large"
                                onClick={handleLogin}
                                disabled={loading}
                                sx={{ mt: 2 }}
                            >
                                Sign In
                            </Button>
                        </Box>
                    </Paper>
                </Grid>
            )}
        </Layout>
    );
};

export default RedirectLoginPage;
