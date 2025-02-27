package command

import (
	"fmt"
	"purge-manager/utils"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the daemon",
	PreRun: func(cmd *cobra.Command, args []string) {
		initRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Stopping purge-manager ...")
		utils.StopProcess()
		fmt.Println("OK")
	},
}
