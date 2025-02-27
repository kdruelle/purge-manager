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
    _ "github.com/go-sql-driver/mysql"
)


func (t* Table) Init() (error) {
    return t.init(nil)
}

func (t * Table) init(p * Table) (error){
    var err error
    for i, _ := range t.Related {
        err = t.Related[i].init(t)
        if err != nil {
            return err
        }
    }
    if !t.SkipIndexes {
        err = t.CheckIndexes(p)
    }
    return err
}


