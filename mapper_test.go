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
	Id                   int32    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Username             string   `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	UserAddr             string   `protobuf:"bytes,3,opt,name=user_addr,proto3" json:"user_addr,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// ordinary struct def
type DemoNoJSONTag struct {
	Id       int32
	Username string
	UserAddr string
}

func mockRowsToSQLRows(mockRows *sqlmock.Rows) *sql.Rows {
	db, mock, _ := sqlmock.New()
	mock.ExpectQuery("select").WillReturnRows(mockRows)
	rows, _ := db.Query("select")
	return rows
}

func mockTestEmptyRow() *sql.Rows {
	mockrows := sqlmock.NewRows([]string{"id", "username", "user_addr"})
	return mockRowsToSQLRows(mockrows)
}

func mockTestOneRow() *sql.Rows {
	mockrows := sqlmock.NewRows([]string{"id", "username", "user_addr"}).
		AddRow(1, "name1", "addr1")
	return mockRowsToSQLRows(mockrows)
}

func mockTestMultiRows() *sql.Rows {
	mockrows := sqlmock.NewRows([]string{"id", "username", "user_addr"}).
		AddRow(1, "name1", "addr1").
		AddRow(2, "name2", "addr2").
		AddRow(3, "name3", "addr3")
	return mockRowsToSQLRows(mockrows)
}

func TestParseJsonStruct(t *testing.T) {
	demo := &DemoProto{}
	jsonMapper := parseStructMemberNames(reflect.TypeOf(demo))
	for jsonname, id := range jsonMapper {
		if id == 0 {
			assert.Equal(t, "id", jsonname, "id not match")
		} else if id == 1 {
			assert.Equal(t, "username", jsonname, "username not match")
		} else if id == 2 {
			assert.Equal(t, "user_addr", jsonname, "user_addr not match")
		} else {
			t.Errorf("unexpected id %v jsonname %v", id, jsonname)
		}
	}
}

func TestParseNoJsonStruct(t *testing.T) {
	demo := &DemoNoJSONTag{}
	jsonMapper := parseStructMemberNames(reflect.TypeOf(demo))
	// t.Errorf("%#v", jsonMapper)
	for jsonname, id := range jsonMapper {
		if id == 0 {
			assert.Equal(t, "id", jsonname, "id not match")
		} else if id == 1 {
			assert.Equal(t, "username", jsonname, "username not match")
		} else if id == 2 {
			assert.Equal(t, "user_addr", jsonname, "user_addr not match")
		} else {
			t.Errorf("unexpected id %v jsonname %v", id, jsonname)
		}
	}
}

func TestParseEmptyRowToStruct(t *testing.T) {
	onerow := mockTestEmptyRow()
	inStruct := DemoProto{}
	err := MapRowsToPointer(onerow, &inStruct)
	assert.Equal(t, err, errors.New(noSQLResult), "supposed to be no result")
}

func TestParseEmptyRowToSlice(t *testing.T) {
	onerow := mockTestEmptyRow()
	inSlices := []*DemoProto{}
	err := MapRowsToPointer(onerow, &inSlices)
	assert.Equal(t, err, errors.New(noSQLResult), "supposed to be no result")
}

func TestParseOneRowToStruct(t *testing.T) {
	onerow := mockTestOneRow()
	inStruct := DemoProto{}
	err := MapRowsToPointer(onerow, &inStruct)
	assert.Nil(t, err)
	expected := DemoProto{
		Id:       1,
		Username: "name1",
		UserAddr: "addr1",
	}
	assert.Equal(t, expected, inStruct, "result not match")
}

func TestParseOneRowToSlice(t *testing.T) {
	onerow := mockTestOneRow()
	inSlices := []*DemoProto{}
	err := MapRowsToPointer(onerow, &inSlices)
	assert.Nil(t, err)
	for _, v := range inSlices {
		expected := &DemoProto{
			Id:       1,
			Username: "name1",
			UserAddr: "addr1",
		}
		assert.Equal(t, expected, v, "result not match")
	}
}

func TestParseMutliRowsToStruct(t *testing.T) {
	multiRows := mockTestMultiRows()
	inStruct := DemoProto{}
	err := MapRowsToPointer(multiRows, &inStruct)
	assert.Nil(t, err)
	expected := DemoProto{
		Id:       1,
		Username: "name1",
		UserAddr: "addr1",
	}
	assert.Equal(t, expected, inStruct, "result not match")
}

func TestParseMutliRowsToSlice(t *testing.T) {
	multiRows := mockTestMultiRows()
	inSlices := []*DemoProto{}
	err := MapRowsToPointer(multiRows, &inSlices)
	assert.Nil(t, err)
	expected := []*DemoProto{
		&DemoProto{
			Id:       1,
			Username: "name1",
			UserAddr: "addr1",
		},
		&DemoProto{
			Id:       2,
			Username: "name2",
			UserAddr: "addr2",
		},
		&DemoProto{
			Id:       3,
			Username: "name3",
			UserAddr: "addr3",
		},
	}
	assert.Equal(t, expected, inSlices, "result not match")
}
