# MySQL Mapper to Golang Struct

---

When querying data from mysql, we used lots of code to match `SQL Result` to `Golang struct`. This module is intended for convenient map between them.

## Method

```Golang
// MapRowsToPointer : Map Mysql result rows to certain pointer
// Parameter 1 - rows : result from MySQL
// Parameter 2 - pointer : must be the pointer of the struct or slice
func MapRowsToPointer(rows *sql.Rows, pointer interface{}) error ```

## Example Code

```Golang
// db is from sql.Open()
db,err := sql.Open("mysql","user:password@tcp(ip:port)/db")
if err != nil {
    return err
}
rows, err := db.Query("select id,username,user_addr from table")
if err != nil {
    return err
}
defer rows.Close()

type A struct{}
// pointer cannot be null, either is pointed to struct or slice of point to struct
pointer1 := A{}
err = mysqlmapper.MatchRowsToPointer(rows, &pointer1)

pointer2 := []*A{}
err = mysqlmapper.MatchRowsToPointer(rows, &pointer2)
```
