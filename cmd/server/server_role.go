package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/models"
)

const (
	serverRoleFlag    = "server-role"
	isBuildServerFlag = "is-build-server"
)

var serverRoleUsage = fmt.Sprintf("Server role (%s): 'deployment' requires another usable build server", strings.Join(models.ServerRoles, "|"))

// addServerRoleFlag registers the --server-role flag.
func addServerRoleFlag(cmd *cobra.Command) {
	cmd.Flags().String(serverRoleFlag, "", serverRoleUsage)
}

// addDeprecatedIsBuildServerFlag registers the legacy --is-build-server flag, kept for backwards compatibility.
func addDeprecatedIsBuildServerFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(isBuildServerFlag, false, "Deprecated: use --server-role build (true) or --server-role both (false)")
	_ = cmd.Flags().MarkDeprecated(isBuildServerFlag, "use --server-role build|both instead")
}

// resolveServerRole returns the server_role value to send, derived from --server-role or the
// deprecated --is-build-server flag. The bool result reports whether a role was requested.
func resolveServerRole(cmd *cobra.Command) (string, bool, error) {
	roleChanged := cmd.Flags().Changed(serverRoleFlag)
	legacyChanged := cmd.Flags().Lookup(isBuildServerFlag) != nil && cmd.Flags().Changed(isBuildServerFlag)

	if roleChanged && legacyChanged {
		return "", false, fmt.Errorf("--%s and --%s cannot be used together; use --%s only", serverRoleFlag, isBuildServerFlag, serverRoleFlag)
	}

	if roleChanged {
		role, _ := cmd.Flags().GetString(serverRoleFlag)
		role, err := normalizeServerRole(role)
		if err != nil {
			return "", false, err
		}
		return role, true, nil
	}

	if legacyChanged {
		isBuildServer, _ := cmd.Flags().GetBool(isBuildServerFlag)
		if isBuildServer {
			return models.ServerRoleBuild, true, nil
		}
		return models.ServerRoleBoth, true, nil
	}

	return "", false, nil
}

func normalizeServerRole(role string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(role))
	if !slices.Contains(models.ServerRoles, normalized) {
		return "", fmt.Errorf("invalid --%s %q: must be one of %s", serverRoleFlag, role, strings.Join(models.ServerRoles, ", "))
	}
	return normalized, nil
}
