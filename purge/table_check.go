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
	"errors"
	"fmt"
	"log"
	"purge-manager/utils"
)

func (t *Table) CheckIndexes(p *Table) error {
	query := fmt.Sprintf("EXPLAIN SELECT t.* FROM %s AS t", t.Schema.Name)
	if p != nil {
		query = fmt.Sprintf("%s INNER JOIN %s AS p ON %s", query, p.Schema.Name, t.Join)
	}
	if t.Condition != "" {
		query = fmt.Sprintf("%s WHERE %s", query, t.Condition)
	}
	rows, err := t.Conn.Query(query)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	if p != nil {
		for rows.Next() {
			row := utils.ScanRow(rows)
			if string(row["table"].([]byte)) == "p" {
				if row["possible_keys"] == nil {
					log.Printf("%s : NO_INDEX_USED : %s", t.Schema.Name, query)
					return errors.New("NO_INDEX_USED")
				}
				break
			}
		}
	}
	return err
}
