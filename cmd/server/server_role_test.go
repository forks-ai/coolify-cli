package server

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseServerRoleFlags(t *testing.T, cmd *cobra.Command, args ...string) (string, bool, error) {
	t.Helper()
	cmd.Flags().SetOutput(io.Discard)
	require.NoError(t, cmd.Flags().Parse(args))
	return resolveServerRole(cmd)
}

func TestResolveServerRole_UpdateCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantRole   string
		wantSet    bool
		wantErrMsg string
	}{
		{name: "no flags", args: nil, wantRole: "", wantSet: false},
		{name: "deployment", args: []string{"--server-role", "deployment"}, wantRole: "deployment", wantSet: true},
		{name: "build", args: []string{"--server-role", "build"}, wantRole: "build", wantSet: true},
		{name: "both", args: []string{"--server-role=both"}, wantRole: "both", wantSet: true},
		{name: "normalizes case and whitespace", args: []string{"--server-role", " Build "}, wantRole: "build", wantSet: true},
		{name: "invalid role", args: []string{"--server-role", "worker"}, wantErrMsg: `invalid --server-role "worker": must be one of deployment, build, both`},
		{name: "empty role", args: []string{"--server-role", ""}, wantErrMsg: `invalid --server-role "": must be one of deployment, build, both`},
		{name: "legacy true maps to build", args: []string{"--is-build-server"}, wantRole: "build", wantSet: true},
		{name: "legacy explicit true maps to build", args: []string{"--is-build-server=true"}, wantRole: "build", wantSet: true},
		{name: "legacy false maps to both", args: []string{"--is-build-server=false"}, wantRole: "both", wantSet: true},
		{name: "both flags conflict", args: []string{"--server-role", "build", "--is-build-server"}, wantErrMsg: "--server-role and --is-build-server cannot be used together"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, set, err := parseServerRoleFlags(t, NewUpdateCommand(), tt.args...)
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.False(t, set)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantRole, role)
			assert.Equal(t, tt.wantSet, set)
		})
	}
}

func TestResolveServerRole_AddCommand(t *testing.T) {
	role, set, err := parseServerRoleFlags(t, NewAddCommand())
	require.NoError(t, err)
	assert.False(t, set)
	assert.Empty(t, role, "add must omit server_role so the API default (both) applies")

	role, set, err = parseServerRoleFlags(t, NewAddCommand(), "--server-role", "build")
	require.NoError(t, err)
	assert.True(t, set)
	assert.Equal(t, "build", role)

	_, _, err = parseServerRoleFlags(t, NewAddCommand(), "--server-role", "invalid")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of deployment, build, both")
}

func TestServerRoleFlags_Registration(t *testing.T) {
	add := NewAddCommand()
	require.NotNil(t, add.Flags().Lookup("server-role"))
	assert.Nil(t, add.Flags().Lookup("is-build-server"), "add never had --is-build-server")

	update := NewUpdateCommand()
	roleFlag := update.Flags().Lookup("server-role")
	require.NotNil(t, roleFlag)
	assert.Contains(t, roleFlag.Usage, "deployment|build|both")

	legacyFlag := update.Flags().Lookup("is-build-server")
	require.NotNil(t, legacyFlag)
	assert.NotEmpty(t, legacyFlag.Deprecated, "--is-build-server must be marked deprecated")
	assert.Contains(t, legacyFlag.Deprecated, "--server-role")
}

func TestServerCommands_RejectInvalidRoleBeforeCallingAPI(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *cobra.Command
		args    []string
		wantErr string
	}{
		{
			name:    "update invalid role",
			cmd:     NewUpdateCommand(),
			args:    []string{"server-uuid", "--server-role", "worker"},
			wantErr: "must be one of deployment, build, both",
		},
		{
			name:    "update conflicting flags",
			cmd:     NewUpdateCommand(),
			args:    []string{"server-uuid", "--server-role", "both", "--is-build-server=false"},
			wantErr: "cannot be used together",
		},
		{
			name:    "add invalid role",
			cmd:     NewAddCommand(),
			args:    []string{"name", "10.0.0.1", "key-uuid", "--server-role", "worker"},
			wantErr: "must be one of deployment, build, both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.SetArgs(tt.args)
			tt.cmd.SetOut(io.Discard)
			tt.cmd.SetErr(io.Discard)
			tt.cmd.SilenceUsage = true
			tt.cmd.SilenceErrors = true

			err := tt.cmd.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
