## Go SQLite Structure Generator

### Usage
Run the command:
```
go run main.go <sqlite-db-path> <output-file>
```


### Features
- Connects to SQLite database.
- Reads schema and generates Go structs.
- Writes structs and initialization function to file.

### Output
- Go code files and packages with constant for database tables.
- Function to initialize and return a database struct.

Primarily targeting Pocketbase users, as I find it hard to keep track of dynamic types during development