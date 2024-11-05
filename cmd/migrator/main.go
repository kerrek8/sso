package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storageHost, storagePort, migrationsPath, migrationsTable, storageUser, storagePassword string

	flag.StringVar(&storageHost, "storage-host", "", "host to storage")
	flag.StringVar(&storageUser, "storage-user", "", "user to storage")
	flag.StringVar(&storagePassword, "storage-password", "", "password to storage")
	flag.StringVar(&storagePort, "storage-port", "", "port to storage")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()
	if storageHost == "" || storagePort == "" || storageUser == "" || storagePassword == "" {

		panic("storage-host, storage-port migrations-path are required")
	}
	m, err := migrate.New("file://"+migrationsPath, fmt.Sprintf("postgres://%s:%s@%s:%s/sso?sslmode=disable&x-migrations-table=%s", storageUser, storagePassword, storageHost, storagePort, migrationsTable))
	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no changes")
			return
		}
		panic(err)
	}
}
