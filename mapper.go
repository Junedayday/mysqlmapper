package mysqlmapper

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const noSQLResult = "no return data from MySQL"

// MapRowsToPointer : Map Mysql result rows to certain pointer
// Parameter 1 - rows : result from MySQL
// Parameter 2 - pointer : must be the pointer of the struct or slice
func MapRowsToPointer(rows *sql.Rows, pointer interface{}) error {
	var nameMapperID map[string]int
	var oneRowType reflect.Type
	var isStruct bool
	pointerVal := reflect.ValueOf(pointer)

	if reflect.ValueOf(pointer).Kind() != reflect.Ptr {
		return fmt.Errorf("input type must be ptr")
	} else if reflect.ValueOf(pointer).Elem().Kind() == reflect.Struct {
		isStruct = true
		nameMapperID = parseStructMemberNames(pointerVal.Type())
		oneRowType = pointerVal.Type().Elem()
	} else if reflect.ValueOf(pointer).Elem().Kind() == reflect.Slice {
		isStruct = false
		nameMapperID = parseStructMemberNames(pointerVal.Elem().Type().Elem())
		oneRowType = pointerVal.Elem().Type().Elem().Elem()
	} else {
		return fmt.Errorf("input pointer must point to struct or slice,but %v", reflect.ValueOf(pointer).Elem().Kind())
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	indexMatch := matchColsToStruct(columns, nameMapperID)

	noRowReturn := true
	oneRow := make([]interface{}, len(columns))
	for j := 0; j < len(oneRow); j++ {
		oneRow[j] = new(sql.RawBytes)
	}
	for rows.Next() {
		noRowReturn = false
		// Instance a struct for scanning
		oneRowStruct := reflect.New(oneRowType).Interface()
		s := reflect.ValueOf(oneRowStruct).Elem()
		for src, target := range indexMatch {
			oneRow[src] = s.Field(target).Addr().Interface()
			rows.Scan(oneRow...)
			oneRow[src] = new(sql.RawBytes)
		}
		if isStruct {
			reflect.ValueOf(pointer).Elem().Set(reflect.ValueOf(oneRowStruct).Elem())
			return nil
		}
		pointerVal.Elem().Set(reflect.Append(pointerVal.Elem(), reflect.ValueOf(oneRowStruct)))
	}

	if noRowReturn {
		return errors.New(noSQLResult)
	}

	// set the slice pointerVal to the input pointer
	reflect.ValueOf(pointer).Elem().Set(pointerVal.Elem())

	return nil
}

// IsEmptyError check the sql result if it is error
func IsEmptyError(err error) bool {
	if err != nil && err.Error() == noSQLResult {
		return true
	}
	return false
}

func parseStructMemberNames(pt reflect.Type) map[string]int {
	el := pt.Elem()
	nameMapperID := make(map[string]int, el.NumField())
	for i := 0; i < el.NumField(); i++ {
		// parse definition in "json:..." in Tags
		// if not find, then transfer name to snake type
		js := el.Field(i).Tag.Get("json")
		if js == "" {
			nameMapperID[snakeString(el.Field(i).Name)] = i
			continue
		} else if js == "-" {
			// "-" means ignore the member value
			continue
		}
		// json tag in proto3 has ","
		// like json:"user_addr,omitempty"`
		comma := strings.Index(js, ",")
		if comma != -1 {
			nameMapperID[js[:comma]] = i
		} else {
			nameMapperID[js] = i
		}
	}
	return nameMapperID
}

func matchColsToStruct(columns []string, mp map[string]int) map[int]int {
	structToColumn := make(map[int]int)
	for index, name := range columns {
		if id, ok := mp[name]; ok {
			structToColumn[index] = id
		}
	}
	return structToColumn
}

func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:]))
}
