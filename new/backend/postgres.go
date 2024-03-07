package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {

	DbHandlers["postgres"] = DbHandler(func(dbName, dbUser, dbPassword, dbAddress, dbTimeZone string) (db *gorm.DB, err error) {

		defaultPort := 5432

		host, port, found := strings.Cut(dbAddress, ":")

		if !found {
			log.Println("Address '" + dbAddress + "' does not use <IP>:<PORT> format, assuming default Postgres port " + strconv.Itoa(defaultPort))
			host = dbAddress
			port = strconv.Itoa(defaultPort)
		}

		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s sslmode=disable TimeZone=%s",
			host,
			port,
			dbUser,
			dbPassword,
			dbTimeZone,
		)
		log.Println(dsn)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		stmt := fmt.Sprintf("SELECT * FROM pg_database WHERE datname = '%s';", dbName)
		rs := db.Raw(stmt)
		if rs.Error != nil {
			log.Fatal(rs.Error)
		}

		// if not create it
		var rec = make(map[string]interface{})
		rs.Find(rec)
		if len(rec) == 0 {

			log.Println("Database does not exist, creating it")
			stmt := fmt.Sprintf("CREATE DATABASE %s;", dbName)
			if rs := db.Exec(stmt); rs.Error != nil {
				log.Fatal(rs.Error)
			}

			// close db connection
			sql, err := db.DB()
			defer func() {
				_ = sql.Close()
			}()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Println("Database exists, using it")
		}

		dsn = dsn + " dbname=" + dbName

		log.Println(dsn)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		return
	})
}
