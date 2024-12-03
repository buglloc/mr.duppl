package main

import (
	"fmt"
	"os"

	"github.com/buglloc/mr.duppl/software/cmd/mr.duppl/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
