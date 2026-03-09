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
	"CityState":   "City/State",
}

var _amexOutputColumns = []string{"Date", "Description", "Amount", "Reference", "Category", "CityState"}

func LoadAmex(r io.Reader) (csvtable.Table, error) {
	return csvtable.Parse(r, _amexColumns, _amexOutputColumns)
}

type AmexParser struct{}

func (AmexParser) Parse(table csvtable.Table, row []string) (transaction.Transaction, error) {
	date, err := formatDate(row[table.Column("Date")])
	if err != nil {
		return transaction.Transaction{}, err
	}
	desc := row[table.Column("Description")]
	ref := row[table.Column("Reference")]
	category := row[table.Column("Category")]
	cityState := row[table.Column("CityState")]

	amount, err := transaction.ParseAmount(row[table.Column("Amount")])
	if err != nil {
		return transaction.Transaction{}, err
	}

	// Reference values are wrapped in single quotes, e.g. '320260630725324753'.
	ref = strings.Trim(ref, "'")

	city, state := parseCityState(cityState)
	location := formatLocation(city, state)
	desc = cleanDescription(desc)

	return transaction.Transaction{
		ID:          ref,
		Date:        date,
		Description: desc,
		Location:    location,
		Amount:      amount,
		Category:    category,
	}, nil
}

// parseCityState splits Amex's "CITY\nSTATE" format.
func parseCityState(s string) (city, state string) {
	city, state, _ = strings.Cut(s, "\n")
	return strings.TrimSpace(city), strings.TrimSpace(state)
}

func formatLocation(city, state string) string {
	switch {
	case city != "" && state != "":
		return city + ", " + state
	case city != "":
		return city
	case state != "":
		return state
	}
	return ""
}

// Amex descriptions are a 42-character fixed-width field:
// characters 0-19 are the merchant name, 20-41 are location (city + state).
// Shorter descriptions don't have embedded location.
const _amexDescWidth = 42
const _amexMerchantWidth = 20

func cleanDescription(desc string) string {
	if len(desc) == _amexDescWidth {
		return strings.TrimSpace(desc[:_amexMerchantWidth])
	}
	return strings.TrimSpace(desc)
}
