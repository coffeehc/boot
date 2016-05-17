package dbservice

import (
	"flag"
	"fmt"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type DatabaseConfig struct {
	Host         string `yaml:"host"`
	Port         string `yaml:port`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	DatabaseName string `yaml:"databaseName"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxOpenConns int    `yaml:"maxOpenConns"`
}

type DBService interface {
	GetDB() *gorm.DB
}

func NewDBService(config *DatabaseConfig, initDB func(db *gorm.DB)) (DBService, error) {
	db, err := newDB(config.User, config.Password, config.Host, config.Port, config.DatabaseName)
	if err != nil {
		return nil, err
	}
	if base.IsDevModule() {
		db.LogMode(true)
		db = db.Debug()
	}
	db.DB().SetMaxIdleConns(config.MaxIdleConns)
	db.DB().SetMaxOpenConns(config.MaxOpenConns)
	initDB(db)
	return &_DBService{
		config: config,
		db:     db,
	}, nil
}

type _DBService struct {
	config *DatabaseConfig
	db     *gorm.DB
}

func (this *_DBService) GetDB() *gorm.DB {
	return this.db
}

func newDB(userName, password, host, port, dbname string) (*gorm.DB, error) {
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
