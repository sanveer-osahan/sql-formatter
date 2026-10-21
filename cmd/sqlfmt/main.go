package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sanveerosahan/sql-formatter/internal/formatter"
)

func main() {
	sql := flag.String("sql", "", "SQL text to format")
	flag.Parse()

	if *sql == "" {
		flag.Usage()
		os.Exit(2)
	}

	out, err := formatter.Format(*sql)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

// sql -> tokens (SELECT (projections) FROM ()) -> "\n".join(tokens)
