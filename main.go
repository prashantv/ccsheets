package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/prashantv/ccsheets/csvtable"
	"github.com/prashantv/ccsheets/provider"
	"github.com/prashantv/ccsheets/transaction"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	providerFlag := flag.String("provider", "", "provider name: chase, amex, citi (auto-detected from filename if omitted)")
	outputFlag := flag.String("output", "table", "output format: table, json")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.PrintDefaults()
		return fmt.Errorf("usage: ccsheets [flags] <csv-file>...")
	}

	var txns []transaction.Transaction
	for _, csvPath := range flag.Args() {
		fileTxns, err := loadFile(csvPath, *providerFlag)
		if err != nil {
			return fmt.Errorf("%s: %w", csvPath, err)
		}
		txns = append(txns, fileTxns...)
	}

	switch *outputFlag {
	case "json":
		printJSON(txns)
	case "table":
		printTable(txns)
	default:
		return fmt.Errorf("unknown output format %q", *outputFlag)
	}

	return nil
}

func loadFile(csvPath, providerName string) ([]transaction.Transaction, error) {
	if providerName == "" {
		var err error
		providerName, err = detectProvider(csvPath)
		if err != nil {
			return nil, err
		}
	}

	loader, parser, err := providerFor(providerName)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(csvPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	table, err := loader(f)
	if err != nil {
		return nil, fmt.Errorf("loading CSV: %w", err)
	}

	return transaction.ParseAll(table, parser)
}

func detectProvider(path string) (string, error) {
	base := strings.ToLower(filepath.Base(path))
	switch {
	case strings.HasPrefix(base, "chase"):
		return "chase", nil
	case strings.HasPrefix(base, "amex"):
		return "amex", nil
	case strings.HasPrefix(base, "citi"):
		return "citi", nil
	}
	return "", fmt.Errorf("unrecognized filename prefix")
}

func providerFor(name string) (func(io.Reader) (csvtable.Table, error), transaction.Parser, error) {
	switch strings.ToLower(name) {
	case "chase":
		return provider.LoadChase, provider.ChaseParser{}, nil
	case "amex":
		return provider.LoadAmex, provider.AmexParser{}, nil
	case "citi":
		return provider.LoadCiti, provider.CitiParser{}, nil
	}
	return nil, nil, fmt.Errorf("unknown provider %q (valid: chase, amex, citi)", name)
}

func printTable(txns []transaction.Transaction) {
	rows := make([][]string, len(txns)+1)
	rows[0] = []string{"ID", "DATE", "DESCRIPTION", "AMOUNT", "CATEGORY"}
	for i, txn := range txns {
		rows[i+1] = []string{txn.ID, txn.Date, txn.Description, txn.Amount.String(), txn.Category}
	}

	// Compute max width per column.
	widths := make([]int, len(rows[0]))
	for _, row := range rows {
		for col, val := range row {
			if len(val) > widths[col] {
				widths[col] = len(val)
			}
		}
	}

	for _, row := range rows {
		for col, val := range row {
			if col > 0 {
				fmt.Print("  ")
			}
			if col < len(row)-1 {
				fmt.Printf("%-*s", widths[col], val)
			} else {
				fmt.Print(val)
			}
		}
		fmt.Println()
	}
}

func printJSON(txns []transaction.Transaction) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	enc.Encode(txns)
}
