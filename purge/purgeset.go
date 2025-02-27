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

package purge

import (
	"database/sql"
	"purge-manager/config"
	"purge-manager/utils"
	"time"

	"github.com/go-sql-driver/mysql"
	//"fmt"
)

type PurgeSet struct {
	Name  string   `hcl:",key"`
	Table []*Table `hcl:"table"`
	Conn  *sql.DB
}

func NewPurgeSet(c config.PurgeSetConfig, softDelete bool) *PurgeSet {
	p := &PurgeSet{
		Name: c.Name,
	}

	config := mysql.NewConfig()

	config.User = c.Database.User
	config.Passwd = c.Database.Password
	config.Net = "tcp"
	config.Addr = c.Database.Host
	config.DBName = c.Database.Schema
	config.Timeout = 20 * time.Second

	conn, err := sql.Open("mysql", config.FormatDSN())
	utils.ExitOnError(err)
	p.Conn = conn
	for _, tc := range c.Table {
		t := NewTable(tc, conn, c.Database.Schema, softDelete)
		p.Table = append(p.Table, t)
	}
	return p
}

func (p *PurgeSet) init() {
	for i := range p.Table {
		err := p.Table[i].Init()
		utils.ExitOnError(err)
	}
}

func (p *PurgeSet) Start() {

	p.init()
	defer p.Conn.Close()

	for _, table := range p.Table {
		table.Purge()
	}
}

func (p *PurgeSet) Count() {
	p.init()
	defer p.Conn.Close()
	for _, table := range p.Table {
		table.Count()
	}
}
