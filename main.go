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
    "github.com/cosiner/flag"
    "os"
)


var(
    version     string
    buildTime   string
)


type Args struct {
    Config   string  `names:"-c, --config"      usage:"Configuration file"              default:"/etc/purge-manager.conf"`
    Count    bool    `names:"-n, --count"       usage:"Print count of rows to delete for the root tables."`
    Verbose  bool    `names:"-v, --verbose"     usage:"Print logs on screen."`
    Debug    bool    `names:"-d, --debug"       usage:"Increase log verbosity and print logs on screen."`
    Soft     bool    `names:"-s, --soft"        usage:"Do not perform the delete operation. This option will leave temporaries table on database"`
    Help     bool    `names:"-h, --help"        usage:"Print this help message."`
}

func (a * Args) Metadata() (map[string]flag.Flag) {
    var(
        usage   = "Purge Manager"
        version = "vertion: v" + version + "\n" + "built: " + buildTime
        desc    = "Purge Manager is a tool for purging old database record"
    )
    return map[string]flag.Flag {
        "" : {
            Usage:      usage,
            Version:    version,
            Desc:       desc,
        },
    }
}

var args Args

func main() {

    flag.NewFlagSet(flag.Flag{}).ParseStruct(&args, os.Args...)

    initLog()

    config, err := ParseConfig()
    exitOnError(err)

    for _, p := range config.Purges {
        p.Start()
    }
}



