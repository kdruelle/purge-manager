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
    "time"
    "database/sql"
    "github.com/go-sql-driver/mysql"
    //_ "github.com/siddontang/go-mysql/driver"
)

type DatabaseConfig struct {
    Host        string
    User        string
    Password    string
    Schema      string
    Dsn         string
}

type PurgeSet struct {
    Name            string              `hcl:",key"`
    Table           []Table             `hcl:"table"`
    Database        DatabaseConfig
}



func (p * PurgeSet) Start() {

    config := mysql.NewConfig()

    config.User     = p.Database.User
    config.Passwd   = p.Database.Password
    config.Net      = "tcp"
    config.Addr     = p.Database.Host
    config.DBName   = p.Database.Schema
    config.Timeout  = 20 * time.Second


    conn, err := sql.Open("mysql", config.FormatDSN())
    exitOnError(err)
    defer conn.Close()

    for i, _ := range p.Table {
        p.Table[i].Conn   = conn
        p.Table[i].Schema = p.Database.Schema
        err = p.Table[i].Init()
        exitOnError(err)
    }
    for _, table := range p.Table {
        if args.Count {
            table.Count()
            continue
        }
        table.Purge()
    }
}


