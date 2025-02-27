package command

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   string
	BuildTime string
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "PrintVersion informations",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Purge Manager v%s-%s-%s    %s", Version, runtime.GOOS, runtime.GOARCH, BuildTime)
		fmt.Println()
	},
}
