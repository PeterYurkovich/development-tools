package update

import (
	"github.com/spf13/cobra"
)

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update and scale up observability components",
	Long:  "Commands for updating and scaling up observability components",
}
