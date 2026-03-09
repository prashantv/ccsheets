package provider

import (
	"io"
	"strings"

	"github.com/prashantv/ccsheets/csvtable"
	"github.com/prashantv/ccsheets/transaction"
)

var _ transaction.Parser = AmexParser{}

// _amexColumns maps output names to CSV header names.
// These columns exist in both Amex Platinum and Amex Blue exports.
var _amexColumns = csvtable.ColumnMap{
	"Date":        "Date",
	"Description": "Description",
	"Amount":      "Amount",
	"Reference":   "Reference",
	"Category":    "Category",
}

var _amexOutputColumns = []string{"Date", "Description", "Amount", "Reference", "Category"}

func LoadAmex(r io.Reader) (csvtable.Table, error) {
	return csvtable.Parse(r, _amexColumns, _amexOutputColumns)
}

type AmexParser struct{}

func (AmexParser) Parse(table csvtable.Table, row []string) (transaction.Transaction, error) {
	date := row[table.Column("Date")]
	desc := row[table.Column("Description")]
	amount := row[table.Column("Amount")]
	ref := row[table.Column("Reference")]
	category := row[table.Column("Category")]

	// Reference values are wrapped in single quotes, e.g. '320260630725324753'.
	ref = strings.Trim(ref, "'")

	return transaction.Transaction{
		ID:          ref,
		Date:        date,
		Description: desc,
		Amount:      amount,
		Category:    category,
	}, nil
}
