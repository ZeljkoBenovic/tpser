package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	cmd := newRootCmd()

	if err := cmd.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v", err)

		os.Exit(1)
	}
}
