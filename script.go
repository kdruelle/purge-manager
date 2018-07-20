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
    "reflect"
    "gopkg.in/olebedev/go-duktape.v3"
    "time"
)

type Script struct {
    name string
    ctx  *duktape.Context
}

func NewScript(name string) (*Script) {
    s := &Script{
        name:   name,
        ctx:    duktape.New(),
    }
    err := s.ctx.PevalFile(s.name)
    if err != nil {
        panic(err)
    }
    s.ctx.Pop()
    return s
}

func (s * Script) Close() {
    s.ctx.DestroyHeap()
}

func (s * Script) pushArray(a []interface{}) {
    a_idx := s.ctx.PushArray()
    for i, v := range a {
        s.pushVal(v)
        s.ctx.PutPropIndex(a_idx, uint(i))
    }
}

func (s * Script) pushObject(obj map[string]interface{}) {
    obj_idx := s.ctx.PushObject()
    for name, val := range obj {
        s.pushVal(val)
        s.ctx.PutPropString(obj_idx, name)
    }
}

func (s * Script) pushVal(val interface{}) {
    switch val.(type) {
    case string:
        s.ctx.PushString(val.(string))
        return
    case []byte:
        s.ctx.PushString(string(val.([]byte)))
        return
    case int64:
        s.ctx.PushUint(uint(val.(int64)))
        return
    case time.Time:
        s.ctx.PushString(val.(time.Time).Format("2006-01-02 15:04:05"))
        return
    case nil:
        s.ctx.PushNull()
        return
    }
    rt := reflect.TypeOf(val)
    switch rt.Kind() {
    case reflect.Slice:
        s.pushArray(InterfaceSlice(val))
    case reflect.Array:
        s.pushArray(InterfaceSlice(val))
    case reflect.Map:
        s.pushObject(val.(map[string]interface{}))
    default:
        s.ctx.PushNull()
    }
}

func (s * Script) Call(function string, args ...interface{}) string {
        s.ctx.PushGlobalObject()
        s.ctx.GetPropString(-1, function)

        for _, arg := range args {
            s.pushVal(arg)
        }

        if s.ctx.Pcall(len(args)) != 0 {
            r := s.ctx.SafeToString(-1)
            panic(r)
        }

        r := s.ctx.GetString(-1)
        s.ctx.Pop2()
        return r
}

func InterfaceSlice(slice interface{}) []interface{} {
    s := reflect.ValueOf(slice)
    if s.Kind() != reflect.Slice {
        panic("InterfaceSlice() given a non-slice type")
    }

    ret := make([]interface{}, s.Len())

    for i:=0; i<s.Len(); i++ {
        ret[i] = s.Index(i).Interface()
    }

    return ret
}

