package gormdb

import (
	"context"
	"sync"

	"gorm.io/gorm"
)

// T is a type representing the database of mappings
type GormDb[T comparable] struct {
	config *Config
	db     map[T]*gorm.DB
	mutex  sync.Mutex
}

func NewGormDb[T comparable](config *Config) *GormDb[T] {
	return &GormDb[T]{
		config: config,
	}
}

func (o *GormDb[T]) Start(ctx context.Context) error {
	return nil
}

func (o *GormDb[T]) Stop(ctx context.Context) error {
	return nil
}
