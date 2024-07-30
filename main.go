package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	strcase "github.com/iancoleman/strcase"
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
		log.Fatal("Usage: go run main.go <sqlite-database-file> <output-file>")
	}
	databaseFile := os.Args[1]
	outputFile := os.Args[2]

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

	// Write the database structure to a file as go code
	f, err := os.Create(outputFile)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	_, err = f.WriteString("package main\n\n")

	if err != nil {
		log.Fatal(err)
	}

	_, err = f.WriteString("type Database struct {\n")
	if err != nil {
		log.Fatal(err)

	}
	for _, table := range database.Tables {
		if strings.HasPrefix(table.Name, "_") {
			continue
		}
		_, err = f.WriteString(fmt.Sprintf("\t%s %s\n", strcase.ToCamel(table.Name), strcase.ToCamel(table.Name)))
		if err != nil {
			log.Fatal(err)
		}
	}
	_, err = f.WriteString("}\n\n")

	if err != nil {
		log.Fatal(err)
	}

	for _, table := range database.Tables {
		if strings.HasPrefix(table.Name, "_") {
			continue
		}
		_, err = f.WriteString(fmt.Sprintf("type %s struct {\n", strcase.ToCamel(table.Name)))
		if err != nil {
			log.Fatal(err)
		}
		for _, column := range table.Columns {
			_, err = f.WriteString(fmt.Sprintf("\t%s %s\n", strcase.ToCamel(column.Name), "string"))
			if err != nil {
				log.Fatal(err)
			}
		}
		_, err = f.WriteString("}\n\n")
		if err != nil {
			log.Fatal(err)
		}
	}

	// fill up each table struct
	filleUpStr := ""

	for _, table := range database.Tables {
		if strings.HasPrefix(table.Name, "_") {
			continue
		}

		filleUpStr += fmt.Sprintf("%s := %s{\n", table.Name, strcase.ToCamel(table.Name))

		for _, column := range table.Columns {
			filleUpStr += fmt.Sprintf("\t%s: \"%s\",\n", strcase.ToCamel(column.Name), column.Name)
		}

		filleUpStr += "}\n\n"

	}

	fillUpTableStr := "db:= Database{\n"

	for _, table := range database.Tables {
		if strings.HasPrefix(table.Name, "_") {
			continue
		}

		fillUpTableStr += fmt.Sprintf("\t%s: %s,\n", strcase.ToCamel(table.Name), table.Name)
	}

	fillUpTableStr += "}\n"

	_, err = f.WriteString("func InitDataTypes() Database {\n")

	if err != nil {
		log.Fatal(err)
	}
	_, err = f.WriteString(filleUpStr)
	_, err = f.WriteString(fillUpTableStr)
	_, err = f.WriteString("return db\n")
	_, err = f.WriteString("\n}")

	if err != nil {
		log.Fatal(err)
	}
}
