package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/prashantv/ccsheets/csvtable"
	"github.com/prashantv/ccsheets/provider"
	"github.com/prashantv/ccsheets/sheet"
	"github.com/prashantv/ccsheets/transaction"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: ccsheets <command> [flags] <files>...\n\ncommands: parse, upload")
	}

	switch os.Args[1] {
	case "parse":
		return runParse(os.Args[2:])
	case "upload":
		return runUpload(os.Args[2:])
	default:
		return fmt.Errorf("unknown command %q (valid: parse, upload)", os.Args[1])
	}
}

func runParse(args []string) error {
	fs := flag.NewFlagSet("ccsheets parse", flag.ContinueOnError)
	providerFlag := fs.String("provider", "", "provider name: chase, amex, citi (auto-detected from filename if omitted)")
	outputFlag := fs.String("output", "table", "output format: table, json")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		fs.PrintDefaults()
		return fmt.Errorf("usage: ccsheets parse [flags] <csv-file>...")
	}

	txns, err := loadFiles(fs.Args(), *providerFlag)
	if err != nil {
		return err
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

func runUpload(args []string) error {
	fs := flag.NewFlagSet("ccsheets upload", flag.ContinueOnError)
	providerFlag := fs.String("provider", "", "provider name: chase, amex, citi (auto-detected from filename if omitted)")
	spreadsheetID := fs.String("spreadsheet-id", os.Getenv("CCSHEETS_SPREADSHEET_ID"), "Google Sheets spreadsheet ID (env: CCSHEETS_SPREADSHEET_ID)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *spreadsheetID == "" {
		return fmt.Errorf("spreadsheet ID required: use -spreadsheet-id or set CCSHEETS_SPREADSHEET_ID")
	}
	if fs.NArg() == 0 {
		fs.PrintDefaults()
		return fmt.Errorf("usage: ccsheets upload [flags] <csv-file>...")
	}

	txns, err := loadFiles(fs.Args(), *providerFlag)
	if err != nil {
		return err
	}

	ctx := context.Background()
	client, err := sheet.NewClient(ctx, *spreadsheetID, "" /* sheetName */)
	if err != nil {
		return err
	}

	added, err := client.Upload(ctx, txns)
	if err != nil {
		return fmt.Errorf("uploading: %w", err)
	}

	fmt.Printf("uploaded %d new transactions (%d total, %d already existed)\n", added, len(txns), len(txns)-added)
	return nil
}

func loadFiles(paths []string, providerName string) ([]transaction.Transaction, error) {
	var txns []transaction.Transaction
	for _, csvPath := range paths {
		fileTxns, err := loadFile(csvPath, providerName)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", csvPath, err)
		}
		txns = append(txns, fileTxns...)
	}
	return txns, nil
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

	txns, err := transaction.ParseAll(table, parser)
	if err != nil {
		return nil, err
	}

	prov, account := providerAccount(providerName, csvPath)
	for i := range txns {
		txns[i].Provider = prov
		txns[i].Account = account
	}
	return txns, nil
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

// providerAccount extracts the display provider name and account identifier from the filename.
// Chase filenames: Chase3501_Activity... → ("Chase", "3501")
// Amex filenames: AmexPlat-... / AmexBlue-... → ("Amex", "Plat" / "Blue")
// Citi filenames: Citi-... → ("Citi", "")
func providerAccount(providerName, path string) (string, string) {
	base := filepath.Base(path)
	switch strings.ToLower(providerName) {
	case "chase":
		// Extract digits between "Chase" and "_".
		if i := strings.Index(base, "_"); i > 5 {
			return "Chase", base[5:i]
		}
		return "Chase", ""
	case "amex":
		// Extract card type between "Amex" and "-".
		if i := strings.Index(base, "-"); i > 4 {
			return "Amex", base[4:i]
		}
		return "Amex", ""
	case "citi":
		return "Citi", ""
	}
	return providerName, ""
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
	rows[0] = []string{"ID", "PROVIDER", "ACCOUNT", "DATE", "DESCRIPTION", "LOCATION", "AMOUNT", "CATEGORY"}
	for i, txn := range txns {
		rows[i+1] = []string{txn.ID, txn.Provider, txn.Account, txn.Date, txn.Description, txn.Location, txn.Amount.String(), txn.Category}
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
