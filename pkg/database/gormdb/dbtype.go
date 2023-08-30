package gormdb

type DbTypeGetDatabase interface {
	GetDatabase() string
}

type DbTypeGetTables interface {
	GetTables() []interface{}
}

type DbType interface {
	DbTypeGetDatabase
	DbTypeGetTables
}
