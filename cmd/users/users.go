package users

import (
	"github.com/spf13/cobra"
)

var UsersCmd = &cobra.Command{
	Use:   "users",
	Short: "Manage test users and RBAC",
	Long:  "Create and manage test users with htpasswd authentication and RBAC permissions",
}
