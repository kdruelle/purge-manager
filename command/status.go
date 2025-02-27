package command

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Stop the daemon",
	PreRun: func(cmd *cobra.Command, args []string) {
		initRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("purge-manager is not runing")
		os.Exit(0)
	},
}
