package command

import (
	"purge-manager/purge"

	"github.com/spf13/cobra"
)

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Perform purge one-shot",
	PreRun: func(cmd *cobra.Command, args []string) {
		initRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		purge.StartPurge(soft)
	},
}
