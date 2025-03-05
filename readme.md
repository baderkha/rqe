# RQE - Restful Query Expression

[![Go Report Card](https://goreportcard.com/badge/github.com/baderkha/rqe)](https://goreportcard.com/report/github.com/baderkha/rqe)
[![License](https://img.shields.io/github/license/baderkha/rqe)](https://github.com/baderkha/rqe/blob/main/LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/baderkha/rqe.svg)](https://pkg.go.dev/github.com/baderkha/rqe)

**RQE (Readable Query Expression)** is a **lightweight and powerful query expression parser** for Go that translates **human-readable** query strings into **structured SQL queries** with placeholders and arguments.

---

## 🚀 Features
✅ **SQL Query Generation** – Converts human-friendly filters into SQL  
✅ **Supports Logical Operators** – `AND`, `OR`, `IN`, `BETWEEN`  
✅ **Custom Column Validation** – Prevents invalid column names and sql injection
✅ **Multi-Value Expressions** – Supports arrays inside `IN` clauses  
✅ **Auto-Formatted SQL** – Generates clean and readable SQL queries  
✅ **Error Handling with Context** – Returns structured errors with line numbers

---

## 📦 Installation

Install RQE via `go get`:

```sh
go get github.com/baderkha/rqe
```

Import it in your Go project:

```go
import "github.com/baderkha/rqe"
```

---

## 🛠 Usage

### **Basic Example**
```go
package main

import (
	"fmt"
	"github.com/baderkha/rqe"
)

func main() {
	filter := `name eq "John" and age gte 25 or (city eq "New York" and status in ["active", "pending"])`

	validateCol := func(col string) bool {
		return map[string]bool{"name": true, "age": true, "city": true, "status": true}[col]
	}

	query, err := rqe.Parse(filter, validateCol)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Formatted SQL:\n", query.SQL)
	fmt.Println("Arguments:", query.Args)
}
```

### **🔹 Output**
```sql
name = ?
AND age >= ?
OR (
  city = ?
  AND status IN ( ?, ? )
)
```
```go
["John", 25, "New York", "active", "pending"]
```

---

## 🏗 Query Syntax

| Operator   | Meaning       | Example               | SQL Equivalent |
|------------|--------------|-----------------------|---------------|
| `eq`       | Equal        | `age eq 30`           | `age = ?`     |
| `ne`       | Not Equal    | `status ne "banned"`  | `status <> ?` |
| `lt`       | Less Than    | `price lt 100`        | `price < ?`   |
| `lte`      | Less or Equal | `score lte 50`       | `score <= ?`  |
| `gt`       | Greater Than | `rating gt 4.5`      | `rating > ?`  |
| `gte`      | Greater or Equal | `salary gte 5000` | `salary >= ?` |
| `in`       | Multiple Values | `color in ["red","blue"]` | `color IN (?, ?)` |
| `between`  | Range Check  | `age between [18 65]`  | `age BETWEEN ? AND ?` |

### **Logical Operators**
- **AND** – `name eq "Alice" and age gte 21`
- **OR** – `status eq "active" or status eq "pending"`
- **Parentheses** – `( age gte 18 and age lte 65 )`

---

## 🔥 Error Handling

RQE provides structured errors with **line and column numbers**.

```go
query, err := rqe.Parse(`name eq "John" and age eq and city eq "New York"`, validateCol)
if err != nil {
	if parseErr, ok := err.(rqe.ParseError); ok {
		line, col := parseErr.Position()
		fmt.Printf("Error: %s at line %d, column %d\n", parseErr.Error(), line, col)
	} else {
		fmt.Println("Unknown error:", err)
	}
}
```

### **🔹 Example Error Output**
```
Error: expected a valid value for column 'age' at line 1, column 20
```

---

## 💡 Contributing

We welcome contributions! If you find a bug or want to add new features, feel free to **open an issue or a pull request**.

1. **Fork the repository**
2. **Create a new branch** (`git checkout -b feature-branch`)
3. **Commit changes** (`git commit -m "Added new feature"`)
4. **Push to GitHub** (`git push origin feature-branch`)
5. **Create a Pull Request**

---

## 📄 License

This project is licensed under the **MIT License**. See [LICENSE](https://github.com/baderkha/rqe/blob/main/LICENSE) for details.

---

## 📢 Contact & Support

- GitHub: [github.com/baderkha/rqe](https://github.com/baderkha/rqe)
- Issues: [github.com/baderkha/rqe/issues](https://github.com/baderkha/rqe/issues)

Enjoy using **RQE**! 🎉🚀

