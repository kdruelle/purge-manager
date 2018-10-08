// +build !windows

/******************************************************************************
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
    "os"
    "time"
    "syscall"
    "github.com/mitchellh/go-ps"
)

func StopProcess() {
    currentProcess, _ := ps.FindProcess(os.Getpid())
    processes, _ := ps.Processes()
    for _, p := range processes {
        if p.Pid() != currentProcess.Pid() && p.Executable() == currentProcess.Executable() {
            syscall.Kill(p.Pid(), syscall.SIGINT)
            active, _ := IsPidActive(p.Pid())
            for active {
                time.Sleep(time.Second)
                active, _ = IsPidActive(p.Pid())
            }
        }
    }
}

