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
    "database/sql"
    "github.com/go-sql-driver/mysql"
    "reflect"
)


func ScanRow(rows * sql.Rows) (row map[string]interface{}) {
    columns, _ := rows.ColumnTypes()
    dest    := make([]interface{}, len(columns))
    row      = make(map[string]interface{})
    for i, col := range columns {
        row[col.Name()] = reflect.New(col.ScanType()).Interface()
        dest[i] = row[col.Name()]

    }

    rows.Scan(dest...)

    for _, col := range columns {
        row[col.Name()] = reflect.Indirect(reflect.ValueOf(row[col.Name()])).Interface()
        switch row[col.Name()].(type) {
        case sql.NullInt64:
            if row[col.Name()].(sql.NullInt64).Valid {
                row[col.Name()] = row[col.Name()].(sql.NullInt64).Int64
                continue
            }
            row[col.Name()] = nil
        case sql.RawBytes:
            row[col.Name()] = []byte(row[col.Name()].(sql.RawBytes))
        case mysql.NullTime:
            if row[col.Name()].(mysql.NullTime).Valid {
                row[col.Name()] = row[col.Name()].(mysql.NullTime).Time
                continue
            }
            row[col.Name()] = nil
        }
    }
    return
}




