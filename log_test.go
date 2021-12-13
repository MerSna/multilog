package multilog

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func getDB() (*sql.DB, error) {
	path := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", "root", "123", "127.0.0.1:3306", "demo")

	DataBase, err := sql.Open("mysql", path)
	if err != nil {
		return nil, err
	}

	// 验证连接
	if err := DataBase.Ping(); err != nil {
		return nil, err
	} else {
		return DataBase, nil
	}
}
func TestPrintDefault(t *testing.T) {
	l := G()
	db, err := getDB()
	if err != nil {
		l.Error(err)
		log.Println()
		return
	}
	l.DBConf = &DatabaseConfig{
		DB:        db,
		TableName: "demo",
	}
	l.Info(123)
}
