package main

import "fmt"
import "database/sql"
import _ "github.com/go-sql-driver/mysql"

func main() {
	/*
		var count int
		db, _ := sql.Open("mysql", "user_name:password@tcp($ip:3305)/?charset=utf8&&autocommit=true")
		sql_str := "select count(*) from test.t3"
		db.QueryRow(sql_str).Scan(&count)
		fmt.Println(count)
	*/
	// [map[name:kun id:5] map[id:6 name:guangkun]]

	var row string
	db, _ := sql.Open("mysql", "user_name:password@tcp($ip:3305)/?charset=utf8&&autocommit=true")
	sql_str := "select name from test.t3 where id = 6"
	db.QueryRow(sql_str).Scan(&row)
	fmt.Println(row)
}
