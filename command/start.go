package command

import (
	"os"
	"purge-manager/purge"
	"purge-manager/utils"
	"runtime"

	"github.com/pingcap/log"
	"github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

func init() {
	startCmd.PersistentFlags().BoolVarP(&soft, "soft", "s", false, "Do not perform the delete operation. This option will leave temporaries table on database")
	startCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print logs on screen.")
	startCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon",
	PreRun: func(cmd *cobra.Command, args []string) {
		initRun()
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Start")
		if verbose || debug || runtime.GOOS == "windows" {
			purge.StartPurgeCron(soft)
			os.Exit(0)
		}
		context := new(daemon.Context)
		child, _ := context.Reborn()
		if child == nil {
			defer context.Release()
			if utils.IsProcessRunning() {
				log.Fatal("another instance of the app is already running, exiting")
			}
			purge.StartPurgeCron(soft)
		}
		os.Exit(0)
	},
}
