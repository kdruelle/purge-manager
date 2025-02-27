package command

import (
	"fmt"
	"os"
	"purge-manager/config"
	"purge-manager/utils"

	"github.com/spf13/cobra"
)

var (
	configFile string
	help       bool

	soft    bool
	verbose bool
	debug   bool
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&help, "help", "h", false, "Print this help message.")
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/purge-manager.conf", "Configuration file.")

	purgeCmd.PersistentFlags().BoolVarP(&soft, "soft", "s", false, "Do not perform the delete operation. This option will leave temporaries table on database")
	purgeCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print logs on screen.")
	purgeCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")

	countCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")

	rootCmd.AddCommand(countCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(purgeCmd)

}

var rootCmd = &cobra.Command{
	Use:   "purge-manager",
	Short: "Purge Manager is a tool for purging old database record",
	Long: `Purge Manager
  Delete periodicaly database old records`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func initRun() {
	err := config.ParseConfig(configFile)
	utils.ExitOnError(err)
	utils.InitLog(verbose, debug)
	if utils.IsProcessRunning() {
		fmt.Println("purge-manager is already runing")
		os.Exit(1)
	}
}
