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
    "fmt"
    "time"
    log "github.com/sirupsen/logrus"
    "database/sql"
    "github.com/go-sql-driver/mysql"
)

const(
    populateLimit   = 500
    deleteLimit     = 200
)


type Table struct {
    Name             string             `hcl:",key"`
    Related          []Table            `hcl:"table"`
    Condition        string
    Parent          string
    Join            string
    Script          string
    Schema           string
    SkipIndexes     bool                `hcl:"skip_indexes"`
    Conn            *sql.DB
}


func (t * Table) Purge() {
    t.purge(nil)
}

func (t * Table) purge(p * Table) {
    populate := t.createPurgeTable(p)
    if populate && t.Script != "" {
        t.populatePurgeScript(p)
    } 
    if populate && t.Script == "" {
        t.populatePurgeTable(p)
    }
    for _, r := range t.Related {
        r.Schema = t.Schema
        r.Conn   = t.Conn
        r.purge(t)
    }
    t.deleteData()
    t.dropPurgeTable()
}

func (t * Table) createPurgeTable(p * Table) bool {
    query := fmt.Sprintf("CREATE TABLE purge_%s LIKE %s", t.Name, t.Name)
    log.Debug(query)
    _, err := t.Conn.Exec(query)
    if err != nil {
        me, ok := err.(*mysql.MySQLError)
        if !ok {
            panic(err)
        }
        switch me.Number {
            case 1050: // table already exista
        default:
            panic(err)
        }
        return false
    }
    return true
}

func (t * Table) populatePurgeTable(p * Table) {
    tt := StartTimeTracker()
    var offset  int64 = 0
    var ra      int64 = 0
    for {
        query := fmt.Sprintf("REPLACE INTO purge_%s SELECT t.* FROM %s AS t", t.Name, t.Name)
        if p != nil {
            query = fmt.Sprintf("%s INNER JOIN purge_%s AS p ON %s", query, p.Name, t.Join)
        }
        if t.Parent != "" {
            query = fmt.Sprintf("%s LEFT OUTER JOIN %s AS p ON %s", query, t.Parent, t.Join)
        }
        if t.Condition != "" {
            query = fmt.Sprintf("%s WHERE %s", query, t.Condition)
        }
        query += fmt.Sprintf(" LIMIT %d,%d", offset, populateLimit)
        log.Debug(query)
        result, err := t.Conn.Exec(query)
        exitOnError(err)
        lra, _ := result.RowsAffected()
        ra += lra
        if lra < populateLimit {
            break
        }
        offset += lra
        time.Sleep(5 * time.Millisecond)
    }
    log.Info(t.Name, " : ", ra, " records prepared for delete in ", tt.ElapsedHuman(), ".")
}

func (t * Table) populatePurgeScript(p * Table) {
    var ra int64 = 0
    script := NewScript(t.Script)
    defer script.Close()

    offset := 0

    for {
        var funcArgs []map[string]interface{}
        query := fmt.Sprintf("SELECT t.* FROM purge_%s AS t LIMIT %d,%d", p.Name, offset, populateLimit)
        log.Debug(query)
        rows, _ := t.Conn.Query(query)
        columns, _ := rows.Columns()
        rawVals := make([][]byte, len(columns))
        dest    := make([]interface{}, len(columns))
        for i, _ := range rawVals {
            dest[i] = &rawVals[i]
        }
        count := 0
        for rows.Next() {
            count++
            row := ScanRow(rows)
            funcArgs = append(funcArgs, row)
        }
        if count > 0 {
            s := script.Call(t.Name, funcArgs)
            log.Debug(fmt.Sprintf("REPLACE INTO purge_%s SELECT * FROM %s WHERE %s", t.Name, t.Name, s))
            result, err := t.Conn.Exec(fmt.Sprintf("REPLACE INTO purge_%s SELECT * FROM %s WHERE %s", t.Name, t.Name, s))
            if err != nil {
                panic(err)
            }
            lra, _ := result.RowsAffected()
            ra += lra
        }
        if count < populateLimit {
            break
        }
        offset += count
    }
    log.Info(t.Name, " : ", ra, " records to delete.")
}

func (t * Table) selectDataToDelete() (data []map[string]interface{}) {
    query := "SELECT "
    for i, pk := range Schemas[t.Name].PKColumns {
        query += fmt.Sprintf("%s", Schemas[t.Name].Columns[pk].Name)
        if i < len(Schemas[t.Name].PKColumns) - 1 {
            query += ", "
        } else {
            query += " "
        }
    }
    query += fmt.Sprintf("FROM purge_%s LIMIT %d", t.Name, deleteLimit)
    log.Debug(query)
    rows, err := t.Conn.Query(query)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        data = append(data, ScanRow(rows))
    }
    return
}

func (t * Table) deleteData() {
    tt := StartTimeTracker()
  
    var countGlobal int64 = 0
    for {

        rowVals := t.selectDataToDelete()
        if rowVals == nil {
            break
        }

        whereQuery := ""
        for _, row := range rowVals {
            if whereQuery != "" {
                whereQuery += " OR "
            }
            indexQuery := ""
            for colname, v := range row {
                if indexQuery != "" {
                    indexQuery += " AND "
                }
                indexQuery += fmt.Sprintf("%s=", colname)
                switch v.(type) {
                case int, uint, int64:
                    indexQuery += fmt.Sprintf("%v", v)
                case time.Time:
                    indexQuery += fmt.Sprintf("'%s'", v.(time.Time).Format("2006-01-02 15:04:05"))
                case []byte:
                    indexQuery += fmt.Sprintf("'%s'", string(v.([]byte)))
                default:
                    indexQuery += fmt.Sprintf("'%v'", v)
                }
            }
            whereQuery += indexQuery
        }
        var logQuery string
        if len(whereQuery) > 100 {
            logQuery = whereQuery[:100] + "..."
        } else {
            logQuery = whereQuery
        }
        log.Debug(fmt.Sprintf("DELETE FROM %s WHERE %s", t.Name, logQuery))

        if args.Soft {
            break
        }

        result, err := t.Conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", t.Name, whereQuery))
        if err != nil {
            panic(err)
        }
        c, _ := result.RowsAffected()
        countGlobal += c
        result, err = t.Conn.Exec(fmt.Sprintf("DELETE FROM purge_%s WHERE %s", t.Name, whereQuery))
        if err != nil {
            panic(err)
        }
        c, _  = result.RowsAffected()
        if c < deleteLimit {
            break;
        }
        time.Sleep(5 * time.Millisecond)
    }
    log.Info(t.Name, " : ", countGlobal, " records was deleted successfuly in ", tt.ElapsedHuman(), ".")
}

func (t * Table) dropPurgeTable() {
    query := fmt.Sprintf("DROP TABLE purge_%s", t.Name)
    log.Debug(query)
    if !args.Soft {
        _, err := t.Conn.Exec(query)
        if err != nil {
            panic(err)
        }
    }
    log.Info(t.Name, " done.")
}

