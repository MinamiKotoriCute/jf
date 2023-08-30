package gormdb

import (
	"fmt"

	"github.com/rotisserie/eris"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (o *GormDb[T]) getDsn(database string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4",
		o.config.User,
		o.config.Password,
		o.config.Host,
		o.config.Password,
		database)
}

func (o *GormDb[T]) connect(dbType T) error {
	databaseName := ""
	if t, ok := any(dbType).(DbTypeGetDatabase); ok {
		databaseName = t.GetDatabase()
	}

	dsn := o.getDsn(databaseName)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return eris.Wrapf(err, "dsn:%s", dsn)
	}

	o.db[dbType] = db
	return nil
}

func (o *GormDb[T]) createDatabaseIfNotExist(databaseName string) error {
	var defaultDbType T
	defaultDb := o.db[defaultDbType]
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", databaseName)
	result := defaultDb.Exec(sql)
	if result.Error != nil {
		eris.Wrapf(result.Error, "sql: %s", sql)
	}

	return nil
}

func (o *GormDb[T]) Connect() error {
	var defaultDbType T
	return o.connect(defaultDbType)
}

// return a gorm.DB or panic
//
// if o.config.AutoMigrate is true:
//
//	will create database if not exist. database name = T.GetDatabase()
//	will auto migrate tables if T.GetTables() is not empty
func (o *GormDb[T]) GetDb(dbType T) *gorm.DB {
	if db, ok := o.db[dbType]; ok {
		return db
	}

	o.mutex.Lock()
	defer o.mutex.Unlock()

	if db, ok := o.db[dbType]; ok {
		return db
	}

	// init db
	if o.config.AutoMigrate {
		databaseName := ""
		if t, ok := any(dbType).(DbTypeGetDatabase); ok {
			databaseName = t.GetDatabase()
		} else {
			panic("databaseName is empty")
		}

		if err := o.createDatabaseIfNotExist(databaseName); err != nil {
			panic(err)
		}
	}

	if err := o.connect(dbType); err != nil {
		panic(err)
	}
	db := o.db[dbType]

	if o.config.AutoMigrate {
		if t, ok := any(dbType).(DbTypeGetTables); ok {
			if tables := t.GetTables(); len(tables) != 0 {
				if err := db.AutoMigrate(tables...); err != nil {
					panic(err)
				}
			}
		}
	}

	return db
}

// return a gorm.DB or panic
func (o *GormDb[T]) Get() *gorm.DB {
	var defaultDbType T
	return o.GetDb(defaultDbType)
}
