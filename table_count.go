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
    "fmt"
    log "github.com/sirupsen/logrus"
)

func (t * Table) Count() {
    t.count(nil)
}

func (t * Table) count(p * Table) {

    if len(t.Related) == 0 && t.Script == "" {
        query := fmt.Sprintf("SELECT count(*) AS count FROM %s AS t", t.Schema.Name)
        if p != nil {
            query = fmt.Sprintf("%s INNER JOIN purge_%s AS p ON %s", query, p.Schema.Name, t.Join)
        }
        if t.Parent != "" {
            query = fmt.Sprintf("%s LEFT OUTER JOIN %s AS p ON %s", query, t.Parent, t.Join)
        }
        if t.Condition != "" {
            query = fmt.Sprintf("%s WHERE %s", query, t.Condition)
        }
        log.Debug(query)
        rows, err := t.Conn.Query(query)
        if err != nil {
            fmt.Println(query)
            panic(err)
        }
        defer rows.Close()
        for rows.Next() {
            rows := ScanRow(rows)
            fmt.Printf("%-30s: %d\n", t.Schema.Name, rows["count"].(int64))
        }
        return
    }


    populate := t.createPurgeTable(p, false)
    defer t.dropPurgeTable()
    var rd int64

    if !populate {
        rd = t.countPurgeTable()
    } else {
        rd = t.populate(p)
    }
    
    fmt.Printf("%-30s: %d\n", t.Schema.Name, rd)

    for _, r := range t.Related {
        r.count(t)
    }
    return


}


