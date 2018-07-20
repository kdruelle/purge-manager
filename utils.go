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
    "time"
    "fmt"
    "os"
    log "github.com/sirupsen/logrus"
)

type TimeTracker struct {
    start   time.Time
}

func StartTimeTracker() (*TimeTracker){
    t := &TimeTracker{
        start: time.Now(),
    }
    return t
}

func (t * TimeTracker) Elapsed() (time.Duration) {
    return time.Since(t.start)
}

func (t * TimeTracker) ElapsedHuman() (string) {
    return t.Elapsed().String()
}

func logTimeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}


func exitOnError(e error) {
    if e == nil {
        return
    }
    if !args.Verbose && !args.Debug {
        fmt.Fprintln(os.Stderr, e)
    }

    log.Error(e)
    os.Exit(1)
}

