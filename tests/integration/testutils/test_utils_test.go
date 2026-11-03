// Copyright 2026 The ThunderID Authors
// SPDX-License-Identifier: Apache-2.0

package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithoutTopLevelKeyDropsTheWholeBlock(t *testing.T) {
	doc := "id: user-1\nsharingPolicies:\n  - childUnits:\n      mode: all\n\n  - otherTreeRoots:\n      mode: all\nattributes:\n  username: alice\n"

	got := WithoutTopLevelKey(doc, "sharingPolicies")

	require.Equal(t, "id: user-1\nattributes:\n  username: alice\n", got)
}

func TestWithoutTopLevelKeyLeavesOtherKeysAlone(t *testing.T) {
	doc := "id: user-1\nattributes:\n  id: nested\n"

	require.Equal(t, "attributes:\n  id: nested\n", WithoutTopLevelKey(doc, "id"))
}

func TestServerPortIgnoresAServerPortTheRunnerDidNotSet(t *testing.T) {
	t.Setenv(TestGroupEnv, "")
	t.Setenv("SERVER_EXTRACTED_HOME", "")
	t.Setenv("SERVER_PORT", "3000")

	require.Equal(t, DefaultServerPort, ServerPort())
}

func TestServerPortUsesTheTestGroupPort(t *testing.T) {
	t.Setenv(TestGroupEnv, "g1")
	t.Setenv("SERVER_EXTRACTED_HOME", "")
	t.Setenv("SERVER_PORT", "8195")

	require.Equal(t, "8195", ServerPort())
}

func TestServerPortUsesThePortTheRunnerPassesToTestPackages(t *testing.T) {
	t.Setenv(TestGroupEnv, "")
	t.Setenv("SERVER_EXTRACTED_HOME", "/tmp/product")
	t.Setenv("SERVER_PORT", "8195")

	require.Equal(t, "8195", ServerPort())
}
