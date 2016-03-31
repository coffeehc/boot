package dbservice

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

func NewDB(userName, password, host, port, dbname string) (*gorm.DB, error) {
	db, err := gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", userName, password, host, port, dbname))
	if err != nil {
		return nil, err
	}
	f := flag.Lookup("devmodule")
	if f.Value.String() == "true" {
		db.Debug()
	}
	mysql := db.DB()
	mysql.SetMaxOpenConns(100)
	mysql.SetMaxIdleConns(5)
	mysql.SetConnMaxLifetime(3 * time.Minute)
	return db, err
}
