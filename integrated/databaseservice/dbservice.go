package databaseservice

import (
	"fmt"
	"github.com/coffeehc/logger"
	"github.com/coffeehc/microserviceboot/base"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"time"
)

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            string        `yaml:port`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DatabaseName    string        `yaml:"databaseName"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
}

func (this *DatabaseConfig) getMaxIdleConns() int {
	if this.MaxIdleConns == 0 {
		this.MaxIdleConns = 5
	}
	return this.MaxIdleConns
}

func (this *DatabaseConfig) getMaxOpenConns() int {
	if this.MaxOpenConns == 0 {
		this.MaxOpenConns = 30
	}
	return this.MaxOpenConns
}

func (this *DatabaseConfig) getConnMaxLifetime() time.Duration {
	if this.ConnMaxLifetime == 0 {
		this.ConnMaxLifetime = time.Second * 5
	}
	return this.ConnMaxLifetime
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
	db.DB().SetMaxIdleConns(config.getMaxIdleConns())
	db.DB().SetMaxOpenConns(config.getMaxOpenConns())
	db.DB().SetConnMaxLifetime(config.getConnMaxLifetime())
	if initDB != nil {
		initDB(db)
	}
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
	logger.Debug("mysql access url is : %s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", userName, "******", host, port, dbname)
	return gorm.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True", userName, password, host, port, dbname))
}
