package mysqlmapper

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// NoSQLResult : no data in sql.Rows
const NoSQLResult = "no return data from MySQL"

// MapRowsToPointer : Map Mysql result rows to certain pointer
// Parameter 1 - rows : result from MySQL
// Parameter 2 - pointer : must be the pointer of the struct or slice
func MapRowsToPointer(rows *sql.Rows, pointer interface{}) error {
	pointerVal := reflect.ValueOf(pointer)
	isStruct := mustBeStructorSlice(pointerVal)

	var nameMapperID map[string]int
	var oneRowType reflect.Type
	if !isStruct {
		nameMapperID = parseStructMemberNames(pointerVal.Elem().Type().Elem())
		oneRowType = pointerVal.Elem().Type().Elem()
	} else {
		nameMapperID = parseStructMemberNames(pointerVal.Type())
		oneRowType = pointerVal.Type()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	indexMatch := matchColsToStruct(columns, nameMapperID)

	noRowReturn := true
	for rows.Next() {
		noRowReturn = false
		oneRowStruct := reflect.New(oneRowType.Elem()).Interface()
		s := reflect.ValueOf(oneRowStruct).Elem()
		onerow := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			id, ok := indexMatch[i]
			if !ok {
				return fmt.Errorf("col %v unmatch the struct", columns[i])
			}
			onerow[i] = s.Field(id).Addr().Interface()
		}
		if err := rows.Scan(onerow...); err != nil {
			return fmt.Errorf("scan problem %v", err)
		}

		if isStruct {
			reflect.ValueOf(pointer).Elem().Set(reflect.ValueOf(oneRowStruct).Elem())
			return nil
		}
		pointerVal.Elem().Set(reflect.Append(pointerVal.Elem(), reflect.ValueOf(oneRowStruct)))
	}

	if noRowReturn {
		return errors.New(NoSQLResult)
	}

	// set the slice pointerVal to the input pointer
	reflect.ValueOf(pointer).Elem().Set(pointerVal.Elem())

	return nil
}

func setNilIf(v *interface{}) {
	*v = nil
}

func mustBeStructorSlice(pointerVal reflect.Value) (isStruct bool) {
	if pointerVal.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("input value must be a pointer,but %v", pointerVal.Kind()))
	} else if pointerVal.Elem().Kind() == reflect.Struct {
		isStruct = true
	} else if pointerVal.Elem().Kind() == reflect.Slice {
		isStruct = false
	} else if pointerVal.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("input pointer must point to struct or slice,but %v", pointerVal.Elem().Kind()))
	}
	return
}

func parseStructMemberNames(pt reflect.Type) map[string]int {
	nameMapperID := make(map[string]int)
	el := pt.Elem()
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
		} else {
			panic(name + " must be found in query struct ")
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
