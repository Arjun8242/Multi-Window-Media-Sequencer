package main

import (
    "bufio"
    "log"
    "net/http"
    "os"
    "strings"
    
    "media-sequencer/internal/database"
    "media-sequencer/internal/handlers"
)

// loadEnv reads key-value pairs from a .env file and sets them in the environment.
func loadEnv(filename string) {
    file, err := os.Open(filename)
    if err != nil {
        return // File does not exist, ignore
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            val := strings.TrimSpace(parts[1])
            if os.Getenv(key) == "" {
                os.Setenv(key, val)
            }
        }
    }
    if err := scanner.Err(); err != nil {
        log.Printf("Warning: error reading %s: %v", filename, err)
    }
}

func main() {
    loadEnv(".env")

    db, err := database.Connect()
    if err != nil {
        log.Fatalf("Database connection failed: %v", err)
    }
    defer db.Close()

    router := handlers.NewRouter(db)

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server listening on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, router))
}
