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
    "github.com/siddontang/go-mysql/schema"
    "github.com/go-sql-driver/mysql"
)

const(
    populateLimit   = 500
    deleteLimit     = 200
)


type Table struct {
    Related          []*Table           `hcl:"table"`
    Condition        string
    Parent          string
    Join            string
    Script          string
    Schema          *schema.Table
    SkipIndexes     bool                `hcl:"skip_indexes"`
    Conn            *sql.DB
}

func NewTable(tc TableConfig, conn * sql.DB, db string) (* Table) {
    t := &Table{
        Conn: conn,
        Condition: tc.Condition,
        Parent: tc.Parent,
        Join: tc.Join,
        Script: tc.Script,
        SkipIndexes: tc.SkipIndexes,
    }
    s, err := schema.NewTableFromSqlDB(conn, db, tc.Name)
    if err != nil {
        return nil
    }
    t.Schema = s
    for _, rtc := range tc.Related {
        rt := NewTable(rtc, conn, db)
        t.Related = append(t.Related, rt)
    }
    return t
}

func (t * Table) Purge() {
    t.purge(nil)
}

func (t * Table) purge(p * Table) {
    populate := t.createPurgeTable(p, false)
    defer t.dropPurgeTable()
    var rd int64
    
    tt := StartTimeTracker()

    if !populate {
        rd = t.countPurgeTable()
    } else {
        rd = t.populate(p)
    }

    if rd == 0 {
        log.Info(t.Schema.Name, " : nothing to do.")
        return
    }

    log.Info(t.Schema.Name, " : ", rd, " records prepared for delete in ", tt.ElapsedHuman(), ".")

    for _, r := range t.Related {
        r.purge(t)
    }
    tt = StartTimeTracker()
    dd := t.deleteData()
    log.Info(t.Schema.Name, " : ", dd, " records was deleted successfuly in ", tt.ElapsedHuman(), ".")
    log.Info(t.Schema.Name, " done.")
}

func (t * Table) createPurgeTable(p * Table, persistant bool) bool {
    temporary := "TEMPORARY"
    if persistant {
        temporary = ""
    }
    query := fmt.Sprintf("CREATE %s TABLE purge_%s LIKE %s", temporary, t.Schema.Name, t.Schema.Name)
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

func (t * Table) populate(p * Table) (int64) {
    if t.Script != "" {
        return t.populatePurgeScript(p)
    }
    return t.populatePurgeTable(p)
}

func (t * Table) populatePurgeTable(p * Table) (int64) {
    var offset  int64 = 0
    var ra      int64 = 0
    for {
        query := fmt.Sprintf("REPLACE INTO purge_%s SELECT t.* FROM %s AS t", t.Schema.Name, t.Schema.Name)
        if p != nil {
            query = fmt.Sprintf("%s INNER JOIN purge_%s AS p ON %s", query, p.Schema.Name, t.Join)
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
    return ra
}

func (t * Table) populatePurgeScript(p * Table) (int64) {
    var ra int64 = 0
    script := NewScript(t.Script)
    defer script.Close()

    offset := 0

    for {
        var funcArgs []map[string]interface{}
        query := fmt.Sprintf("SELECT t.* FROM purge_%s AS t LIMIT %d,%d", p.Schema.Name, offset, populateLimit)
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
            s := script.Call(t.Schema.Name, funcArgs)
            var logQuery string
            if len(s) > 100 {
                logQuery = s[:100] + "..."
            } else {
                logQuery = s
            }
            log.Debug(fmt.Sprintf("REPLACE INTO purge_%s SELECT * FROM %s WHERE %s", t.Schema.Name, t.Schema.Name, logQuery))
            result, err := t.Conn.Exec(fmt.Sprintf("REPLACE INTO purge_%s SELECT * FROM %s WHERE %s", t.Schema.Name, t.Schema.Name, s))
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
    return ra
}

func (t * Table) countPurgeTable() (i int64) {
    query := fmt.Sprintf("SELECT count(*) AS count FROM purge_%s AS t", t.Schema.Name)
    rows, err := t.Conn.Query(query)
    if err != nil {
        panic(err)
    }
    defer rows.Close()
    for rows.Next() {
        rows := ScanRow(rows)
        i = rows["count"].(int64)
    }
    return
}

func (t * Table) selectDataToDelete() (data []map[string]interface{}) {
    query := "SELECT "
    for i, pk := range t.Schema.PKColumns {
        query += fmt.Sprintf("%s", t.Schema.Columns[pk].Name)
        if i < len(t.Schema.PKColumns) - 1 {
            query += ", "
        } else {
            query += " "
        }
    }
    query += fmt.Sprintf("FROM purge_%s LIMIT %d", t.Schema.Name, deleteLimit)
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

func (t * Table) deleteData() (int64) {
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
        log.Debug(fmt.Sprintf("DELETE FROM %s WHERE %s", t.Schema.Name, logQuery))

        if soft {
            break
        }

        result, err := t.Conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE %s", t.Schema.Name, whereQuery))
        if err != nil {
            panic(err)
        }
        c, _ := result.RowsAffected()
        countGlobal += c
        result, err = t.Conn.Exec(fmt.Sprintf("DELETE FROM purge_%s WHERE %s", t.Schema.Name, whereQuery))
        if err != nil {
            panic(err)
        }
        c, _  = result.RowsAffected()
        if c < deleteLimit {
            break;
        }
        time.Sleep(5 * time.Millisecond)
    }
    return countGlobal
}

func (t * Table) dropPurgeTable() {
    query := fmt.Sprintf("DROP TABLE purge_%s", t.Schema.Name)
    log.Debug(query)
    if !soft {
        _, err := t.Conn.Exec(query)
        if err != nil {
            panic(err)
        }
    }
}

