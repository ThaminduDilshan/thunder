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

/**
 * Playwright E2E Test Configuration
 *
 * This configuration sets up test projects for Chromium, Firefox, and Webkit.
 * All projects depend on the `setup` project for authentication.
 *
 * Reports are generated in both HTML and Blob format (for merging).
 *
 * @see https://playwright.dev/docs/test-configuration
 */

import { defineConfig, devices } from "@playwright/test";
import dotenv from "dotenv";
import path from "path";
import { Timeouts } from "./constants/timeouts";

const envPath = path.resolve(__dirname, ".env");
dotenv.config({ path: envPath });

const STORAGE_STATE = path.join(__dirname, "playwright/.auth/console-admin.json");

/**
 * Configure number of workers. Workers parallelize test *files* across all projects (including the
 * three browser projects), so one shared pool fans out chromium/firefox/webkit files simultaneously.
 * Default 6 fits the standard GitHub `ubuntu-latest` runner (4 vCPU / 16GB RAM); tune via
 * PLAYWRIGHT_WORKERS if the workflow ever moves to a larger runner or hits memory pressure.
 */
const WORKERS = process.env.PLAYWRIGHT_WORKERS ? parseInt(process.env.PLAYWRIGHT_WORKERS, 10) : 6;

/**
 * Specs that mutate global, non-partitionable server state and cannot run in parallel with themselves
 * across browsers:
 *   - CORS allowed-origins is a server-wide list; the console does read-modify-write on Save, so
 *     concurrent workers overwrite each other's list.
 *   - MFA suite reconfigures the shared sample-app's auth flow and depends on the one notification
 *     sender pointing at the one mock SMS server; parallel workers step on each other's OTPs.
 * These files are excluded from the fan-out projects below and re-added as a serial chain
 * (chromium -> firefox -> webkit) so at most one worker is executing them at any moment.
 */
const SERIAL_SPECS = [
  "**/settings/cors-allowed-origins.spec.ts",
  "**/sample-app-authentication/sample-app-mfa-login.spec.ts",
];

export default defineConfig({
  /** Directory containing test files */
  testDir: "./tests",

  /** Run tests sequentially to avoid auth conflicts */
  fullyParallel: false,

  /** Fail CI builds if test.only() is accidentally committed */
  forbidOnly: !!process.env.CI,

  /** Retry failed tests (more on CI) */
  retries: process.env.CI ? 2 : 1,

  /** Number of workers for parallel execution */
  workers: WORKERS,

  /** Generate HTML report, Console list, and Blob report for merging */
  reporter: [
    ["html"],
    ["list"],
    ["blob"],
    // Add JSON reporter for better CI integration
    ["json", { outputFile: "test-results/test-results.json" }],
    // Add JUnit reporter for CI systems
    ["junit", { outputFile: "test-results/junit.xml" }],
  ],

  /** Global test timeout */
  timeout: 90000,

  /** Expect timeout for assertions */
  expect: {
    timeout: 10000,
    toHaveScreenshot: {
      maxDiffPixels: 100,
      threshold: 0.2,
    },
  },

  /** Global setup and teardown */
  globalSetup: require.resolve("./global-setup"),
  globalTeardown: require.resolve("./global-teardown"),

  /**
   * Run local dev server before starting the tests.
   * This ensures the server is up before the setup project tries to authenticate.
   */
  webServer: {
    command:
      process.platform === "win32"
        ? "cd ..\\..  && pwsh -File .\\build.ps1 run_backend"
        : "cd ../.. && ./build.sh run_backend",
    url: "https://localhost:8090/health/liveness",
    reuseExistingServer: true,
    ignoreHTTPSErrors: true,
    timeout: 120 * 1000,
  },

  /** Shared settings for all projects */
  use: {
    trace: "retain-on-failure",
    ignoreHTTPSErrors: true,
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    actionTimeout: Timeouts.DEFAULT_ACTION,
    baseURL: process.env.BASE_URL || "https://localhost:8090",
    // Add context options for better reliability
    viewport: { width: 1280, height: 720 },
    userAgent: "Playwright E2E Tests",
    // Collect console logs for debugging
    launchOptions: {
      slowMo: process.env.SLOW_MO ? parseInt(process.env.SLOW_MO) : 0,
    },
  },

  projects: [
    /** Setup project - only runs auth.setup.ts */
    {
      name: "setup",
      testMatch: "**/*.setup.ts",
      use: { ...devices["Desktop Chrome"], ignoreHTTPSErrors: true },
    },

    /** Main browser projects - run parallel-safe specs. Serial specs are excluded here and
     *  re-added below in a per-browser chain. */
    {
      name: "chromium",
      testMatch: "**/*.spec.ts",
      testIgnore: SERIAL_SPECS,
      use: {
        ...devices["Desktop Chrome"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["setup"],
    },

    {
      name: "firefox",
      testMatch: "**/*.spec.ts",
      testIgnore: SERIAL_SPECS,
      use: {
        ...devices["Desktop Firefox"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["setup"],
    },

    {
      name: "webkit",
      testMatch: "**/*.spec.ts",
      testIgnore: SERIAL_SPECS,
      use: {
        ...devices["Desktop Safari"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["setup"],
    },

    /**
     * Serial chain for CORS + MFA specs. Each browser depends on the previous one so Playwright's
     * scheduler runs them one at a time even when 6 workers are otherwise fanning out. If a project
     * in the chain fails, later browsers in the chain are skipped (dependency semantics); the job
     * still fails and per-test retries still apply.
     */
    {
      name: "serial-chromium",
      testMatch: SERIAL_SPECS,
      use: {
        ...devices["Desktop Chrome"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["setup"],
    },
    {
      name: "serial-firefox",
      testMatch: SERIAL_SPECS,
      use: {
        ...devices["Desktop Firefox"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["serial-chromium"],
    },
    {
      name: "serial-webkit",
      testMatch: SERIAL_SPECS,
      use: {
        ...devices["Desktop Safari"],
        storageState: STORAGE_STATE,
      },
      dependencies: ["serial-firefox"],
    },
  ],
});
