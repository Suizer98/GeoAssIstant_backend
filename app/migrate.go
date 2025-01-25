package app

import (
    "log"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/postgres"
    _ "github.com/golang-migrate/migrate/v4/source/file"
    _ "github.com/lib/pq"
)

func (a *App) Migrate() {
    // Create PostgreSQL driver
    driver, err := postgres.WithInstance(a.DB, &postgres.Config{})
    if err != nil {
        log.Fatalf("Error creating PostgreSQL driver: %v", err)
        return
    }

    // Initialize migrations
    m, err := migrate.NewWithDatabaseInstance(
        "file://./migrations/",
        "geoaistore",
        driver,
    )
    if err != nil {
        log.Fatalf("Error initializing migrations: %v", err)
        return
    }

    // Apply migrations
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Fatalf("Error applying migrations: %v", err)
        return
    }

    log.Println("Migrations applied successfully!")
}
