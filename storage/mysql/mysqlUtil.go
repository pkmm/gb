package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

const Name = "zhang"
const dev = "root:glority@/52pkm_cn?charset=utf8"
const prod = "root:zcc123@/sql52pkm_cn?charset=utf8"

type Info struct {
	Fid   string
	Infos string
}

var filename = "tieba.log"
var logFile, _ = os.Create(filename)
var Log = log.New(logFile, "[tieba sign]", log.LstdFlags)

func Get(userId string) *map[string]string {
	db, err := sql.Open("mysql", dev)
	if err != nil {
		Log.Println(err)
	}
	rows, err := db.Query("SELECT kw, fid FROM tiebas where sign_at != ? AND user_id = ? AND fid != -1;", time.Now().Day(), userId)
	if err != nil {
		Log.Println(err)
	}
	var result = make(map[string]string)
	var kw, fid string
	for rows.Next() {
		rows.Scan(&kw, &fid)
		result[kw] = fid
	}
	return &result
}

func Update(signed *map[string]Info, userId string) {
	db, err := sql.Open("mysql", dev)
	if err != nil {
		Log.Println(err)
	}
	today := time.Now().Day()
	for _, infos := range *signed {
		db.Exec("UPDATE tiebas set sign_at = ?, sign_infos = ? where user_id = ? AND fid = ?;", today, infos.Infos, userId, infos.Fid)
	}
}

func SyncFromOffical(current *map[string]string, userId string) {
	db, err := sql.Open("mysql", dev)
	if err != nil {
		Log.Println(err)
	}
	for kw, fid := range *current {
		_, err := db.Exec("insert into tiebas (kw, fid, user_id) values (?,?,?) ON DUPLICATE KEY UPDATE kw = kw;", kw, fid, userId)
		if err != nil {
			Log.Println(err)
		}
	}
}
