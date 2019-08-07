# MySQL Mapper to Golang Struct

---

When querying data from mysql, we used lots of code to match `SQL Result` to `Golang struct`. This module is intended for convenient map between them.

---

## Map Rule

- from: `column names` in SQL selection
- to: `golang struct` definition

> All columns in `from` must be found in golang struct

## Golang Strcut

Golang struct is recommended as the following 3 ways:

1. Generated from protobuf definition
2. Define the `json` in tag field
3. Transfer from struct field name as snake string

```golang

// type 1: protobuf
type A struct {
	Id                   int32    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

// type 2: json tag
type A struct {
	Id                   int32    `json:"id,omitempty"`
}

// type 3: transfer field name to snake string
type A struct {
	Id
}

```

## Method

```golang

// MapRowsToPointer : Map Mysql result rows to certain pointer
// Parameter 1 - rows : result from MySQL
// Parameter 2 - pointer : must be the pointer of the struct or slice
func MapRowsToPointer(rows *sql.Rows, pointer interface{}) error

```

## Example Code

```golang

// db is from sql.Open()
db,err := sql.Open("mysql","user:password@tcp(ip:port)/db")
if err != nil {
    return err
}
// result shouldn't be null. if it is, please use ifnull() in sql
rows, err := db.Query("select id,username,user_addr from table")
if err != nil {
    return err
}
defer rows.Close()

type A struct{}
// pointer cannot be null, either is pointed to struct or slice of point to struct
pointer1 := A{}
err = mysqlmapper.MapRowsToPointer(rows, &pointer1)

pointer2 := []*A{}
err = mysqlmapper.MapRowsToPointer(rows, &pointer2)
```
