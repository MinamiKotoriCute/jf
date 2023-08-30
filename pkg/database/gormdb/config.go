package gormdb

type Config struct {
	Host     string
	Port     int
	User     string
	Password string

	AutoMigrate bool
}
