![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/craftdome/go-pipeline?style=flat-square)

# About

A module for assembling a data pipeline.

# Features

- [x] Go channels & goroutines
- [x] Using generics for custom data
- [x] Chain of units

# Quick start

```go
package main

import (
	"context"
	"fmt"
	"github.com/craftdome/go-pipeline"
)

func main() {
	// Initializing a unit with <string> as input type and <int> as output
	unit := pipeline.NewUnit[string, int]()

	// An action that the unit performs
	unit.OnExecute = func(s string) (int, error) {
		return len(s), nil
	}

	unit.Start()
	
	unit.Input() <- "my_string"
	strLen := <-unit.Output()
	fmt.Println(strLen)

	unit.Stop(context.Background())
}
```
