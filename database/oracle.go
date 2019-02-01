// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// This is heavily inspired by meddler.
// We basically rely on sqlx but use this library to build the missing statements

package database

import (
  "database/sql"
  "fmt"
  "time"

  "reflect"
  "strconv"
  "strings"
  "sync"

  null "gopkg.in/guregu/null.v3"
)

const tagName = "db"

type DB interface {
  Exec(query string, args ...interface{}) (sql.Result, error)
  Query(query string, args ...interface{}) (*sql.Rows, error)
  QueryRow(query string, args ...interface{}) *sql.Row
}

// DatabaseSyntax contains driver specific settings.
type DatabaseSyntax struct {
  Quote               string // the quote character for table and column names
  Placeholder         string // the placeholder style to use in generated queries
  UseReturningToGetID bool   // use PostgreSQL-style RETURNING "ID" instead of calling sql.Result.LastInsertID
}

// MySQL contains database specific options for executing queries in a MySQL database
var MySQLSyntax = &DatabaseSyntax{
  Quote:               "`",
  Placeholder:         "?",
  UseReturningToGetID: false,
}

// PostgreSQL contains database specific options for executing queries in a PostgreSQL database
var PostgreSQLSyntax = &DatabaseSyntax{
  Quote:               `"`,
  Placeholder:         "$1",
  UseReturningToGetID: true,
}

// SQLite contains database specific options for executing queries in a SQLite database
var SQLiteSyntax = &DatabaseSyntax{
  Quote:               `"`,
  Placeholder:         "?",
  UseReturningToGetID: false,
}

var DefaultSyntax = PostgreSQLSyntax

func (d *DatabaseSyntax) quoted(s string) string {
  return d.Quote + s + d.Quote
}

func (d *DatabaseSyntax) placeholder(n int) string {
  return strings.Replace(d.Placeholder, "1", strconv.FormatInt(int64(n), 10), 1)
}

// represents an entry in a struct
type structField struct {
  column string
  index  int
}

// represents all entries from a struct
type structInfo struct {
  columns []string
  fields  map[string]*structField
}

var fieldsCache = make(map[reflect.Type]*structInfo)
var fieldsCacheMutex sync.Mutex

// do some reflection on struct to parse "db" tags
func parseStruct(objectType reflect.Type) (*structInfo, error) {
  // use caching to speed things up
  fieldsCacheMutex.Lock()
  defer fieldsCacheMutex.Unlock()

  if result, present := fieldsCache[objectType]; present {
    return result, nil
  }

  // make sure dst is a non-nil pointer to a struct
  if objectType.Kind() != reflect.Ptr {
    return nil, fmt.Errorf("sqlorcale called with non-pointer destination %v", objectType)
  }

  structType := objectType.Elem()
  if structType.Kind() != reflect.Struct {
    return nil, fmt.Errorf("sqlorcale called with pointer to non-struct %v", objectType)
  }

  // gather the list of fields in the struct
  data := new(structInfo)
  data.fields = make(map[string]*structField)

  for i := 0; i < structType.NumField(); i++ {
    f := structType.Field(i)

    // skip non-exported fields
    if f.PkgPath != "" {
      continue
    }

    // examine the tag for metadata
    tag := strings.Split(f.Tag.Get(tagName), ",")

    // was this field marked for skipping?
    if len(tag) > 0 && tag[0] == "-" {
      continue
    }

    // default to the field name
    name := f.Name

    // the tag can override the field name
    if len(tag) > 0 && tag[0] != "" {
      name = tag[0]
    }

    if name == "id" {
      if f.Type.Kind() == reflect.Ptr {
        return nil, fmt.Errorf("sqlorcale found field %s which is the primary key but is a pointer", f.Name)
      }

      // make sure it is an int of some kind
      switch f.Type.Kind() {
      case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
      case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
      default:
        return nil, fmt.Errorf("meddler found field %s which is marked as the primary key, but is not an integer type", f.Name)
      }
    }

    // prevent duplicates in tags
    if _, present := data.fields[name]; present {
      return nil, fmt.Errorf("sqlorcale found multiple fields for column %s", name)
    }
    data.fields[name] = &structField{
      column: name,
      index:  i,
    }
    data.columns = append(data.columns, name)

  }

  fieldsCache[objectType] = data
  return data, nil

}

// Columns returns a list of column names for its input struct.
func (d *DatabaseSyntax) Columns(src interface{}, includePk bool) ([]string, error) {
  structInfo, err := parseStruct(reflect.TypeOf(src))
  if err != nil {
    return nil, err
  }

  var names []string
  for _, elt := range structInfo.columns {
    if !includePk && elt == "id" {
      continue
    }
    names = append(names, elt)
  }

  return names, nil
}

// Columns using the Default Database type.
func Columns(src interface{}, includePk bool) ([]string, error) {
  return DefaultSyntax.Columns(src, includePk)
}

type StatementData struct {
  Column string
  Value  interface{}
}

// isZero tests whether the incoming value is the default value.
// https://stackoverflow.com/a/27494761/7443104
func isZero(v reflect.Value) bool {
  switch v.Kind() {
  case reflect.Func, reflect.Map, reflect.Slice:
    return v.IsNil()
  case reflect.Array:
    z := true
    for i := 0; i < v.Len(); i++ {
      z = z && isZero(v.Index(i))
    }
    return z
  case reflect.Struct:
    z := true
    for i := 0; i < v.NumField(); i++ {
      if v.Field(i).CanSet() {
        z = z && isZero(v.Field(i))
      }
    }
    return z
  case reflect.Ptr:
    if !v.IsNil() {
      return isZero(reflect.Indirect(v))
    }
  }
  // Compare other types directly:
  z := reflect.Zero(v.Type())
  result := v.Interface() == z.Interface()

  return result
}

// PackStatementData reads a struct and extract necessary data for a query.
// This skips the primary key "id" automatically. No data modification is made.
func (d *DatabaseSyntax) PackStatementData(src interface{}) ([]StatementData, error) {

  var null_string null.String

  statementDatas := []StatementData{}

  // extract column names we are interested in
  columns, err := d.Columns(src, false)
  if err != nil {
    return nil, err
  }

  // extract struct info
  // objectType := reflect.TypeOf(src)
  structVal := reflect.ValueOf(src).Elem()
  // structType := objectType.Elem()
  data, err := parseStruct(reflect.TypeOf(src))

  // structVal := reflect.ValueOf(src).Elem()
  for _, name := range columns {
    field, present := data.fields[name]

    if name == "id" {
      continue
    }

    // field is in tag and current struct
    if present {
      current_value := structVal.Field(field.index)
      if !isZero(current_value) {

        // if current_value.Type() == null.String
        // This is an ugly case, but we want to nicely create JSON
        // and the null package does the job
        if reflect.TypeOf(null_string) == current_value.Type() {
          if current_value.Field(0).Field(1).Bool() == true {
            statementDatas = append(statementDatas, StatementData{
              Column: name,
              Value:  current_value.Field(0).Field(0).String(),
            })

          }
        } else {
          statementDatas = append(statementDatas, StatementData{
            Column: name,
            Value:  current_value.Interface(),
          })

        }

      }

    }

  }
  return statementDatas, nil
}

func PackStatementData(src interface{}) ([]StatementData, error) {
  return DefaultSyntax.PackStatementData(src)

}

// Build an sql statement to insert values
// returns the statement and values
// stmt, values, err := BuildInsertStatement(...)
// var newPk int64
// err := db.QueryRow(stmt, values...).Scan(&newPk)
func (d *DatabaseSyntax) InsertStatement(table string, src interface{}) (string, []interface{}, error) {
  stmtData, err := PackStatementData(src)
  if err != nil {
    return "", nil, err
  }

  // structInfo, err := parseStruct(reflect.TypeOf(src))

  var columns []string
  var placeholders []string
  var values []interface{}
  for _, el := range stmtData {
    if el.Column == "created_at" || el.Column == "updated_at" {
      continue
    }
    columns = append(columns, d.quoted(el.Column))
    placeholders = append(placeholders, d.placeholder(len(placeholders)+1))
    values = append(values, el.Value)
  }

  // set "created_at" and "updated_at"
  current_time := time.Now()
  structInfo, err := parseStruct(reflect.TypeOf(src))
  _, present := structInfo.fields["created_at"]
  if present {
    // we will need to set created_at
    columns = append(columns, d.quoted("created_at"))
    placeholders = append(placeholders, d.placeholder(len(placeholders)+1))
    values = append(values, current_time)
  }
  _, present = structInfo.fields["updated_at"]
  if present {
    // we will need to set updated_at
    columns = append(columns, d.quoted("updated_at"))
    placeholders = append(placeholders, d.placeholder(len(placeholders)+1))
    values = append(values, current_time)
  }

  column_string := strings.Join(columns, ", ")
  placeholder_string := strings.Join(placeholders, ", ")

  stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, column_string, placeholder_string)

  if d.UseReturningToGetID {
    stmt += " RETURNING " + d.quoted("id")
  }
  stmt += ";"

  return stmt, values, nil

}

func InsertStatement(table string, src interface{}) (string, []interface{}, error) {
  return DefaultSyntax.InsertStatement(table, src)

}

func (d *DatabaseSyntax) Insert(db DB, table string, src interface{}) (int64, error) {
  stmt, values, err := InsertStatement(table, src)
  if err != nil {
    return 0, err
  }

  if d.UseReturningToGetID {
    // database returns the last id
    var newPk int64
    err := db.QueryRow(stmt, values...).Scan(&newPk)
    return newPk, err
  } else {
    // we need to ask for the last
    result, err := db.Exec(stmt, values...)
    if err != nil {
      return 0, err
    }

    newPk, err := result.LastInsertId()
    if err != nil {
      return 0, err
    }

    return newPk, nil

  }

}

func Insert(db DB, table string, src interface{}) (int64, error) {
  return DefaultSyntax.Insert(db, table, src)
}

// Build an sql statement to update values
// returns the statement and values
// stmt, values, err := BuildInsertStatement(...)
// var newPk int64
// err := db.QueryRow(stmt, values...).Scan(&newPk)
func (d *DatabaseSyntax) UpdateStatement(table string, id int64, src interface{}) (string, []interface{}, error) {
  data, err := PackStatementData(src)
  if err != nil {
    return "", nil, err
  }
  _ = data

  var values []interface{}
  values = append(values, id)
  var pairs []string
  for _, el := range data {
    if el.Column == "created_at" || el.Column == "updated_at" {
      continue
    }
    pairs = append(pairs, fmt.Sprintf("%s = %s", d.quoted(el.Column), d.placeholder(len(pairs)+2)))
    values = append(values, el.Value)
  }

  structInfo, err := parseStruct(reflect.TypeOf(src))
  _, present := structInfo.fields["updated_at"]
  if present {
    // we will need to set updated_at
    pairs = append(pairs, fmt.Sprintf("%s = %s", d.quoted("updated_at"), d.placeholder(len(pairs)+2)))
    values = append(values, time.Now())
  }

  pairs_string := strings.Join(pairs, ", ")
  stmt := fmt.Sprintf("UPDATE %s SET %s WHERE id = $1;", table, pairs_string)

  return stmt, values, nil

}

func UpdateStatement(table string, id int64, src interface{}) (string, []interface{}, error) {
  return DefaultSyntax.UpdateStatement(table, id, src)
}

func (d *DatabaseSyntax) Update(db DB, table string, id int64, src interface{}) error {
  stmt, values, err := UpdateStatement(table, id, src)
  if err != nil {
    return err
  }

  _, err = db.Exec(stmt, values...)
  return err

}

func Update(db DB, table string, id int64, src interface{}) error {
  return DefaultSyntax.Update(db, table, id, src)
}

func (d *DatabaseSyntax) DeleteStatement(table string, id int64) (string, []interface{}) {
  stmt := fmt.Sprintf("DELETE FROM %s WHERE id = $1;", table)

  var values []interface{}
  values = append(values, id)
  return stmt, values
}

func DeleteStatement(table string, id int64) (string, []interface{}) {
  return DefaultSyntax.DeleteStatement(table, id)
}

func (d *DatabaseSyntax) Delete(db DB, table string, id int64) error {
  stmt, values := DeleteStatement(table, id)
  _, err := db.Exec(stmt, values...)
  return err

}

func Delete(db DB, table string, id int64) error {
  return DefaultSyntax.Delete(db, table, id)
}
