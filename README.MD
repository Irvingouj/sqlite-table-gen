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
- Go code file with structs for database tables.
- Function to initialize and return a database struct.

### Example Output
```go
type Database struct { Users Users }
type Users struct { Id string; Name string }
func InitDataTypes() Database { ... }
```

Primarily targeting Pocketbase users, as I find it hard to keep track of dynamic types during development