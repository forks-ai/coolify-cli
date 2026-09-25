package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coollabsio/coolify-cli/internal/cli"
	"github.com/coollabsio/coolify-cli/internal/models"
	"github.com/coollabsio/coolify-cli/internal/service"
)

// NewAddCommand creates the add command
func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <server_name> <ip_address> <private_key_uuid>",
		Args:  cli.ExactArgs(3, "<server_name> <ip_address> <private_key_uuid>"),
		Short: "Add a server",
		Long: `Add a server to Coolify.

Example:
  coolify server add build-1 10.0.0.10 <private_key_uuid> --server-role build`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			serverRole, _, err := resolveServerRole(cmd)
			if err != nil {
				return err
			}

			// Get API client
			client, err := cli.GetAPIClient(cmd)
			if err != nil {
				return fmt.Errorf("failed to get API client: %w", err)
			}

			// Parse arguments and flags
			name := args[0]
			ip := args[1]
			privateKeyUUID := args[2]
			port, _ := cmd.Flags().GetInt("port")
			user, _ := cmd.Flags().GetString("user")
			validate, _ := cmd.Flags().GetBool("validate")

			// Create request
			req := models.ServerCreateRequest{
				Name:            name,
				IP:              ip,
				Port:            port,
				User:            user,
				PrivateKeyUUID:  privateKeyUUID,
				ServerRole:      serverRole,
				InstantValidate: validate,
			}

			// Use service layer
			serverSvc := service.NewServerService(client)
			response, err := serverSvc.Create(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create server: %w", err)
			}

			if validate {
				fmt.Printf("Server added successfully with uuid %s\n", response.UUID)
			} else {
				fmt.Printf("Server added successfully with uuid %s. Server is not validated. Use 'servers validate %s' to validate the server.\n", response.UUID, response.UUID)
			}

			return nil
		},
	}

	cmd.Flags().IntP("port", "p", 22, "Port")
	cmd.Flags().StringP("user", "u", "root", "User")
	cmd.Flags().Bool("validate", false, "Validate the server")
	addServerRoleFlag(cmd)

	return cmd
}
