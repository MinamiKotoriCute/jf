package gormdb

import "fmt"

type DbTypeGetDatabase interface {
	GetDatabase() string
}

type DbTypeGetTables interface {
	GetTables() []interface{}
}

// refer: https://gorm.io/zh_CN/docs/connecting_to_the_database.html
type DbTypeGetDsn interface {
	GetDsn(config *Config) string
}

func GetMysqlDsn(databaseName string, config *Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		databaseName)
}

func GetPostgresqlDsn(databaseName string, config *Config) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Host,
		config.User,
		config.Password,
		databaseName,
		config.Port)
}
