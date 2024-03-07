package main

import (
	"errors"

	"gorm.io/gorm"
)

type DbHandler func(dbName, dbUser, dbPassword, dbAddress, dbTimeZone string) (db *gorm.DB, err error)

var DbHandlers = make(map[DatabaseType]DbHandler)

func InitDB(c *Config) error {
	handler, ok := DbHandlers[c.DatabaseType]
	if !ok {
		return errors.New("Database type '" + string(c.DatabaseType) + "' is not supported or does not exist")
	}

	var err error
	DB, err = handler(c.DatabaseName, c.DatabaseUsername, c.DatabasePassword, c.DatabaseAddress, c.DatabaseTimezone)
	return err
}
