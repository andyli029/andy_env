package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// var count int
	db, _ := sql.Open("mysql", "user_name:password@tcp($ip:3305)/?charset=utf8&&autocommit=true")
	sql_str := "select * from test.t3 limit 2"
	rows, err := db.Query(sql_str)
	fmt.Println(err)
	cols, err := rows.Columns()
	count := len(cols)

	rltData := make([]map[string]interface{}, 0)
	vals := make([]interface{}, count)
	valPtrs := make([]interface{}, count)

	for rows.Next() {
		for i := 0; i < count; i++ {
			valPtrs[i] = &vals[i]
		}
		rows.Scan(valPtrs...)

		rowData := map[string]interface{}{}
		for i, col := range cols {
			var v interface{}
			if b, ok := vals[i].([]byte); ok {
				v = string(b)
			} else {
				v = vals[i]
			}
			rowData[col] = v
		}

		rltData = append(rltData, rowData)
	}
	rows.Close()

	b, err := json.Marshal(rltData)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(string(b))
}
