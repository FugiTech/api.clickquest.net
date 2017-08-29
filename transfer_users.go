package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	oldDB, err := sql.Open("mysql", "xxx/clickquest")
	if err != nil {
		log.Fatal("sql.Open(oldDB):", err)
	}
	newDB, err := sql.Open("mysql", "xxx/clickquest.net")
	if err != nil {
		log.Fatal("sql.Open(newDB):", err)
	}

	// Support resuming
	var maxID int
	err = newDB.QueryRow("SELECT max(id) FROM users").Scan(&maxID)
	if err != nil {
		log.Fatal("newDB.QueryRow():", err)
	}

	rows, err := oldDB.Query("SELECT id, username, password, clicks, modified, color, ip, admin, `mod`, banned, hardcore, online, totaltime FROM users WHERE id > ? ORDER BY id ASC", maxID)
	if err != nil {
		log.Fatal("oldDB.Query:", err)
	}

	for rows.Next() {
		var (
			id        int
			username  string
			password  string
			clicks    int
			modified  int
			color     string
			ip        string
			admin     bool
			mod       bool
			banned    bool
			hardcore  int
			online    int64
			totaltime int
		)
		err = rows.Scan(&id, &username, &password, &clicks, &modified, &color, &ip, &admin, &mod, &banned, &hardcore, &online, &totaltime)
		if err != nil {
			log.Fatal("rows.Scan():", err)
		}
		_, err := newDB.Exec("INSERT INTO users(id, username, oldpassword, clicks, modified, color, ip, admin, `mod`, banned, hardcore, last_online, totaltime) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)",
			id, username, password, clicks, modified, "#"+color, ip, admin, mod, banned, hardcore, time.Unix(online, 0), totaltime)
		if err != nil {
			log.Fatal("newDB.Exec():", err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal("rows.Err():", err)
	}
}
