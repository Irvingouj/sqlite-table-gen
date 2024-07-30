package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/iancoleman/strcase"
	_ "github.com/mattn/go-sqlite3"
)

type Column struct {
	Name string
	Type string
}

type Table struct {
	Name    string
	Columns []Column
}

type DB struct {
	Tables []Table
}

func getTableNames(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, nil
}

func getColumnsForTable(db *sql.DB, tableName string) ([]Column, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var (
			cid        int
			name       string
			ctype      string
			notnull    int
			dflt_value *string
			pk         int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			return nil, err
		}
		columns = append(columns, Column{Name: name, Type: ctype})
	}
	return columns, nil
}

func readDatabaseStructure(db *sql.DB) (DB, error) {
	var database DB
	tableNames, err := getTableNames(db)
	if err != nil {
		return database, err
	}
	for _, tableName := range tableNames {
		columns, err := getColumnsForTable(db, tableName)
		if err != nil {
			return database, err
		}
		database.Tables = append(database.Tables, Table{Name: tableName, Columns: columns})
	}
	return database, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <sqlite-database-file> <output-folder>")
	}
	databaseFile := os.Args[1]
	outputFolder := os.Args[2]

	db, err := sql.Open("sqlite3", databaseFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	database, err := readDatabaseStructure(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Database Structure:\n")
	for _, table := range database.Tables {
		fmt.Printf("Table: %s\n", table.Name)
		for _, column := range table.Columns {
			fmt.Printf("\tColumn: %s, Type: %s\n", column.Name, column.Type)
		}
	}

	// Ensure output folder exists
	if err := os.MkdirAll(outputFolder, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	for _, table := range database.Tables {
		if strings.HasPrefix(table.Name, "_") {
			continue
		}
		generateCodeForTable(table, outputFolder)
	}
}

func generateCodeForTable(table Table, outputFolder string) {
	tableFolder := fmt.Sprintf("%s/%s", outputFolder, table.Name)
	if err := os.MkdirAll(tableFolder, os.ModePerm); err != nil {
		log.Fatal(err)
		return
	}

	filePath := fmt.Sprintf("%s/%s.go", tableFolder, table.Name)
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()

	toWrite := fmt.Sprintf("package %s\n\n", table.Name)
	toWrite += "var (\n"
	toWrite += fmt.Sprintf("\tTableName = \"%s\"\n", table.Name)
	for _, column := range table.Columns {
		toWrite += fmt.Sprintf("\t%s = \"%s\"\n", strcase.ToCamel(column.Name), column.Name)
	}
	toWrite += ")\n"

	if _, err := file.WriteString(toWrite); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("gofmt", "-w", filePath)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
