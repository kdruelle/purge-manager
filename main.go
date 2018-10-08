/******************************************************************************
**
** This file is part of purge-manager.
**
** (C) 2011 Kevin Druelle <kevin@druelle.info>
**
** This software is free software: you can redistribute it and/or modify
** it under the terms of the GNU General Public License as published by
** the Free Software Foundation, either version 3 of the License, or
** (at your option) any later version.
** 
** This software is distributed in the hope that it will be useful,
** but WITHOUT ANY WARRANTY; without even the implied warranty of
** MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
** GNU General Public License for more details.
** 
** You should have received a copy of the GNU General Public License
** along with this software.  If not, see <http://www.gnu.org/licenses/>.
** 
******************************************************************************/

package main

import(
    "github.com/spf13/cobra"
    log "github.com/sirupsen/logrus"
    "github.com/sevlyar/go-daemon"
    "github.com/gorhill/cronexpr"
    "time"
    "os"
    "fmt"
    "os/signal"
    "syscall"
    "runtime"
    "sync"
)


var(
    configFile  string
    help        bool

    soft        bool
    verbose     bool
    debug       bool

    version     string
    buildTime   string
)

func init() {
    rootCmd.PersistentFlags().BoolVarP(&help, "help", "h", false, "Print this help message.")
    rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "/etc/purge-manager.conf", "Configuration file.")

    startCmd.PersistentFlags().BoolVarP(&soft, "soft", "s", false, "Do not perform the delete operation. This option will leave temporaries table on database")
    startCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print logs on screen.")
    startCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")

    purgeCmd.PersistentFlags().BoolVarP(&soft, "soft", "s", false, "Do not perform the delete operation. This option will leave temporaries table on database")
    purgeCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print logs on screen.")
    purgeCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")

    countCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Increase log verbosity and print logs on screen.")
    
    rootCmd.AddCommand(countCmd)
    rootCmd.AddCommand(startCmd)
    rootCmd.AddCommand(stopCmd)
    rootCmd.AddCommand(statusCmd)
    rootCmd.AddCommand(purgeCmd)
    rootCmd.AddCommand(versionCmd)
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

var countCmd = &cobra.Command{
  Use:   "count",
  Short: "Print count of rows to delete for the root tables.",
  PreRun: func(cmd *cobra.Command, args []string) {
      initRun()
  },
  Run: func(cmd *cobra.Command, args []string) {
      for _, pc := range config.Purges {
          p := NewPurgeSet(pc)
          p.Count()
      }
  },
}

var startCmd = &cobra.Command{
  Use:   "start",
  Short: "Start the daemon",
  PreRun: func(cmd *cobra.Command, args []string) {
      initRun()
  },
  Run: func(cmd *cobra.Command, args []string) {
      log.Info("Start")
      if verbose || debug  || runtime.GOOS == "windows" {
        startPurgeCron()
        os.Exit(0)
      }
      context := new(daemon.Context)
      child, _ := context.Reborn()
      if child == nil {
          defer context.Release()
          if IsProcessRunning() {
              log.Fatal("another instance of the app is already running, exiting")
          } 
          startPurgeCron()
      }
      os.Exit(0)
  },
}

var stopCmd = &cobra.Command{
    Use:   "stop",
    Short: "Stop the daemon",
    PreRun: func(cmd *cobra.Command, args []string) {
        err := ParseConfig()
        exitOnError(err)
        initLog()
    },
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Stopping purge-manager ...")
        StopProcess()
        fmt.Println("OK")
    },
}

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Stop the daemon",
    PreRun: func(cmd *cobra.Command, args []string) {
        err := ParseConfig()
        exitOnError(err)
        initLog()
    },
    Run: func(cmd *cobra.Command, args []string) {
        if IsProcessRunning() {
            fmt.Println("purge-manager is runing")
            os.Exit(0)
        }
        fmt.Println("purge-manager is not runing")
        os.Exit(0)
    },
}

var purgeCmd = &cobra.Command{
    Use:   "purge",
    Short: "Perform purge one-shot",
    PreRun: func(cmd *cobra.Command, args []string) {
      initRun()
    },
    Run: func(cmd *cobra.Command, args []string) {
        startPurge()
    },
}

var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "PrintVersion informations",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("Purge Manager v%s-%s-%s    %s", version, runtime.GOOS, runtime.GOARCH, buildTime)
        fmt.Println()
    },
}

func initRun(){
    err := ParseConfig()
    exitOnError(err)
    initLog()
    if IsProcessRunning() {
        fmt.Println("purge-manager is already runing")
        os.Exit(1)
    }
}

func main() {
    err := rootCmd.Execute()
    exitOnError(err)
}

func startPurge() {
    for _, pc := range config.Purges {
        p := NewPurgeSet(pc)
        tt := StartTimeTracker()
        log.Info("purgeset '", p.Name, "' : start.")
        p.Start()
        log.Info("purgeset '", p.Name, "' : done in ", tt.ElapsedHuman(), ".")
    }
}

func startPurgeCron() {
    sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
    var wg sync.WaitGroup
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    for _, pc := range config.Purges {
        go func() {
            wg.Add(1)
            p := NewPurgeSet(pc)
            cron := pc.Cron
            for {
            expr := cronexpr.MustParse(cron)
            nextTime := expr.Next(time.Now())
            log.Info("purgeset '", p.Name, "' : Next execution at ", nextTime)
            duration := nextTime.Sub(time.Now())
            timer := time.NewTimer(duration)
            select {
            case <-timer.C:
                tt := StartTimeTracker()
                log.Info("purgeset '", p.Name, "' : start.")
                p.Start()
                log.Info("purgeset '", p.Name, "' : done in ", tt.ElapsedHuman(), ".")
            case <-done:
                wg.Done()
                return
            }
        }
        }()
    }
    <-sigs
    close(done)
    wg.Wait()
}



