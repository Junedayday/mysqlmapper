package mysqlmapper

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// proto3 def
type DemoProto struct {
	Id       int32   `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Username string  `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	UserAddr string  `protobuf:"bytes,3,opt,name=user_addr,proto3" json:"user_addr,omitempty"`
	Money    float64 `protobuf:"bytes,4,opt,name=money,proto3" json:"money,omitempty"`
	UnitName string   `protobuf:"bytes,5,opt,name=unit_name,json=unitName,proto3" json:"unit_name,omitempty"`
}

// ordinary struct def
type DemoNoJSONTag struct {
	Id       int32
	Username string
	UserAddr string
	Money    float64
}

type NotMatchColumns struct {
	Id       int32
	Username string
	User     string
	Money    float64
}

func mockRowsToSQLRows(mockRows *sqlmock.Rows) *sql.Rows {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(mockRows)
	rows, _ := db.Query("select")
	return rows
}

func mockRawRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "username", "user_addr", "money"})
}

func mockNRows(n int) *sql.Rows {
	mock := mockRawRow()
	for i := 0; i < n; i++ {
		mock = mock.AddRow(10, "name1", "addr1", 1.32)
	}
	return mockRowsToSQLRows(mock)
}

func mockNotMatch2Rows() *sql.Rows {
	mock := mockRawRow()
	mock = mock.AddRow(10, "", "addr1", 1.32).AddRow(11, "name1", "addr1", 0)
	return mockRowsToSQLRows(mock)
}

func checkProto(in DemoProto) bool {
	expected := DemoProto{
		Id:       10,
		Username: "name1",
		UserAddr: "addr1",
		Money:    1.32,
	}
	return reflect.DeepEqual(in, expected)
}

func checkProtoSlice(in []*DemoProto) bool {
	for _, v := range in {
		if !checkProto(*v) {
			return false
		}
	}
	return true
}

func checkStruct(in DemoNoJSONTag) bool {
	expected := DemoNoJSONTag{
		Id:       10,
		Username: "name1",
		UserAddr: "addr1",
		Money:    1.32,
	}
	return reflect.DeepEqual(in, expected)
}

func checkStructSlice(in []*DemoNoJSONTag) bool {
	for _, v := range in {
		if !checkStruct(*v) {
			return false
		}
	}
	return true
}

func TestParseNil(t *testing.T) {
	mock := sqlmock.NewRows([]string{"id", "username", "unit_name", "money"}).AddRow(10,nil, "addr1", 1.32).AddRow(11, "name1", "addr1", 0)
	re:=mockRowsToSQLRows(mock)
	var inProtoSlice []*DemoProto
	assert.Equal(t, nil,MapRowsToPointer(re, &inProtoSlice))
	expected := DemoProto{
		Id:       10,
		UnitName: "addr1",
		Money:    1.32,
	}
	assert.Equal(t, expected,*inProtoSlice[0])
}

func TestParseEmptyRow(t *testing.T) {
	inProto := DemoProto{}
	assert.Equal(t, MapRowsToPointer(mockNRows(0), &inProto), errors.New(noSQLResult), "supposed to be no result")

	var inProtoSlice []*DemoProto
	assert.Equal(t, MapRowsToPointer(mockNRows(0), &inProtoSlice), errors.New(noSQLResult), "supposed to be no result")

	inStruct := DemoNoJSONTag{}
	assert.Equal(t, MapRowsToPointer(mockNRows(0), &inStruct), errors.New(noSQLResult), "supposed to be no result")

	var inStructSlice []*DemoNoJSONTag
	assert.Equal(t, MapRowsToPointer(mockNRows(0), &inStructSlice), errors.New(noSQLResult), "supposed to be no result")
}

func TestParseOneRow(t *testing.T) {
	inProto := DemoProto{}
	assert.Equal(t, MapRowsToPointer(mockNRows(1), &inProto), nil)
	assert.Equal(t, checkProto(inProto), true)

	var inProtoSlice []*DemoProto
	assert.Equal(t, MapRowsToPointer(mockNRows(1), &inProtoSlice), nil)
	assert.Equal(t, checkProtoSlice(inProtoSlice), true)

	inStruct := DemoNoJSONTag{}
	assert.Equal(t, MapRowsToPointer(mockNRows(1), &inStruct), nil)
	assert.Equal(t, checkStruct(inStruct), true)

	var inStructSlice []*DemoNoJSONTag
	assert.Equal(t, MapRowsToPointer(mockNRows(1), &inStructSlice), nil)
	assert.Equal(t, checkStructSlice(inStructSlice), true)
}

func TestParse100Rows(t *testing.T) {
	var inProtoSlice []*DemoProto
	assert.Equal(t, MapRowsToPointer(mockNRows(100), &inProtoSlice), nil)
	assert.Equal(t, checkProtoSlice(inProtoSlice), true)

	var inStructSlice []*DemoNoJSONTag
	assert.Equal(t, MapRowsToPointer(mockNRows(100), &inStructSlice), nil)
	assert.Equal(t, checkStructSlice(inStructSlice), true)
}

func TestParseNotMatchColumns(t *testing.T) {
	inProto := NotMatchColumns{}
	assert.Equal(t, MapRowsToPointer(mockNRows(1), &inProto), nil)
	expected := NotMatchColumns{
		Id:       10,
		Username: "name1",
		User:     "",
		Money:    1.32,
	}
	assert.Equal(t, inProto, expected)

	var inNotMatchSlice []*NotMatchColumns
	assert.Equal(t, MapRowsToPointer(mockNotMatch2Rows(), &inNotMatchSlice), nil)
	expected1 := NotMatchColumns{
		Id:       10,
		Username: "",
		User:     "",
		Money:    1.32,
	}
	expected2 := NotMatchColumns{
		Id:       11,
		Username: "name1",
		User:     "",
		Money:    0,
	}
	assert.Equal(t, *inNotMatchSlice[0], expected1)
	assert.Equal(t, *inNotMatchSlice[1], expected2)
}

func BenchmarkMapOneRowToPointerForTag(b *testing.B) {
	data := mockNRows(1)
	var inProtoSlice []*DemoProto
	err := MapRowsToPointer(data, &inProtoSlice)
	assert.Equal(b, err, nil)
	assert.Equal(b, checkProtoSlice(inProtoSlice), true)
}

func BenchmarkMapOneRowToPointerForStruct(b *testing.B) {
	data := mockNRows(1)
	var inStructSlice []*DemoNoJSONTag
	err := MapRowsToPointer(data, &inStructSlice)
	assert.Equal(b, err, nil)
	assert.Equal(b, checkStructSlice(inStructSlice), true)
}

func BenchmarkOneRowOfficialMethod(b *testing.B) {
	data := mockNRows(1)
	var inStructSlice []*DemoNoJSONTag
	for data.Next() {
		var d = new(DemoNoJSONTag)
		if err := data.Scan(&d.Id, &d.Username, &d.UserAddr, &d.Money); err != nil {
			b.Fatalf("error in scan %v", err)
		}
		inStructSlice = append(inStructSlice, d)
	}
	assert.Equal(b, checkStructSlice(inStructSlice), true)
}

func BenchmarkMapRowsToPointerForTag(b *testing.B) {
	data := mockNRows(10000)
	var inProtoSlice []*DemoProto
	err := MapRowsToPointer(data, &inProtoSlice)
	assert.Equal(b, err, nil)
	assert.Equal(b, checkProtoSlice(inProtoSlice), true)
}

func BenchmarkMapRowsToPointerForStruct(b *testing.B) {
	data := mockNRows(10000)
	var inStructSlice []*DemoNoJSONTag
	err := MapRowsToPointer(data, &inStructSlice)
	assert.Equal(b, err, nil)
	assert.Equal(b, checkStructSlice(inStructSlice), true)
}

func BenchmarkOfficialMethod(b *testing.B) {
	data := mockNRows(10000)
	var inStructSlice []*DemoNoJSONTag
	for data.Next() {
		var d = new(DemoNoJSONTag)
		if err := data.Scan(&d.Id, &d.Username, &d.UserAddr, &d.Money); err != nil {
			b.Fatalf("error in scan %v", err)
		}
		inStructSlice = append(inStructSlice, d)
	}
	assert.Equal(b, checkStructSlice(inStructSlice), true)
}
