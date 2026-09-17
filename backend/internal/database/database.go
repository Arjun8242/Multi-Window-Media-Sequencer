package database

import (
    "database/sql"
    "fmt"
    "os"
    "path/filepath"
    _ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
    dbUrl := os.Getenv("DATABASE_URL")
    if dbUrl == "" {
        dbUrl = "postgres://postgres:postgres@localhost:5433/mediadb?sslmode=disable"
    }
    
    db, err := sql.Open("postgres", dbUrl)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("could not reach database: %w", err)
    }

    // Attempt to run the schema migration automatically
    schemaPath := filepath.Join("migrations", "schema.sql")
    if _, err := os.Stat(schemaPath); err == nil {
        schemaBytes, err := os.ReadFile(schemaPath)
        if err == nil {
            _, execErr := db.Exec(string(schemaBytes))
            if execErr != nil {
                fmt.Printf("Warning: failed to execute schema.sql automatically: %v\n", execErr)
            } else {
                fmt.Println("Schema initialized/verified successfully.")
            }
        }
    }

    return db, nil
}
