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

import (
    "io/ioutil"
    "github.com/hashicorp/hcl"
)

type Config struct {
    Purges  []PurgeSetConfig     `hcl:"purge"`
}

type DatabaseConfig struct {
    Host        string
    User        string
    Password    string
    Schema      string
    Dsn         string
}

type PurgeSetConfig struct {
    Name       string              `hcl:",key"`
    Table      []TableConfig       `hcl:"table"`
    Cron       string
    Database   DatabaseConfig
}

type TableConfig struct {
    Name             string             `hcl:",key"`
    Related          []TableConfig      `hcl:"table"`
    Condition        string
    Parent          string
    Join            string
    Script          string
    Schema           string
    SkipIndexes     bool                `hcl:"skip_indexes"`
}

var config Config

func ParseConfig() (error) {
    b, err := ioutil.ReadFile(configFile)
    if err != nil {
        return err
    }

    err = hcl.Unmarshal(b, &config)
    return err
}


