package database

import (
	"log/slog"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const Storage = "storage/"

func init() {
	os.Mkdir(Storage, os.ModePerm)
}

func NewDatabaseConnection(location string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(Storage+location+"?_pragma=foreign_keys(1)"), &gorm.Config{SkipDefaultTransaction: true, PrepareStmt: true})
	if err != nil {
		slog.Error("Cannot connect to database", "err", err)
		os.Exit(1)
	}
	return db
}
