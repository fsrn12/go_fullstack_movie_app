package main

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

const (
	sqlFilePath = "database-dump.sql" // Update this to the path of your SQL file
)

func init() {
	// Load the .env file from a custom directory outside current project
	envPath := filepath.Join("..", ".env") // Adjust to actual path
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warning: Could not load .env file at %s: %v", envPath, err)
	}
}

func main() {
	// Get the database connection string from environment variables
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL is not set in environment variables")
	}

	// Open connection using pgx driver
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	log.Println("Connected to the database successfully.")

	// Load and parse SQL file
	sqlStatements, err := loadSQLStatements(sqlFilePath)
	if err != nil {
		log.Fatalf("Error loading SQL file: %v", err)
	}

	// Execute SQL statements
	executeSQLStatements(db, sqlStatements)

	log.Println("SQL script execution completed.")
}

// loadSQLStatements reads a SQL file, splits by semicolon, and removes comments
func loadSQLStatements(filePath string) ([]string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	rawStatements := strings.Split(string(content), ";\n")
	var statements []string

	for _, stmt := range rawStatements {
		cleaned := cleanSQLStatement(stmt)
		if cleaned != "" {
			statements = append(statements, cleaned)
		}
	}

	return statements, nil
}

// cleanSQLStatement removes line comments and trims whitespace
func cleanSQLStatement(stmt string) string {
	var cleanedLines []string
	scanner := bufio.NewScanner(strings.NewReader(stmt))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "#") {
			continue
		}
		cleanedLines = append(cleanedLines, line)
	}
	return strings.TrimSpace(strings.Join(cleanedLines, " "))
}

// executeSQLStatements runs the given SQL statements one-by-one
func executeSQLStatements(db *sql.DB, statements []string) {
	for _, stmt := range statements {
		if _, err := db.ExecContext(context.Background(), stmt); err != nil {
			log.Printf("❌ Failed to execute:\n%v\nError: %v\n", stmt, err)
			continue
		}
		log.Printf("✅ Executed: %.50s...", stmt)
	}
}
