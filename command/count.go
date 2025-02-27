package command

import (
	"purge-manager/config"
	"purge-manager/purge"

	"github.com/spf13/cobra"
)

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Print count of rows to delete for the root tables.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, pc := range config.Purges() {
			p := purge.NewPurgeSet(pc, soft)
			p.Count()
		}
	},
}
