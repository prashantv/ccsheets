package provider

import (
	"io"
	"strings"

	"github.com/prashantv/ccsheets/csvtable"
	"github.com/prashantv/ccsheets/transaction"
)

var _ transaction.Parser = CitiParser{}

var _citiColumns = csvtable.ColumnMap{
	"Date":        "Date",
	"Description": "Description",
	"Debit":       "Debit",
	"Credit":      "Credit",
}

var _citiOutputColumns = []string{"Date", "Description", "Debit", "Credit"}

func LoadCiti(r io.Reader) (csvtable.Table, error) {
	return csvtable.Parse(r, _citiColumns, _citiOutputColumns)
}

type CitiParser struct{}

func (CitiParser) Parse(table csvtable.Table, row []string) (transaction.Transaction, error) {
	date, err := formatDate(row[table.Column("Date")])
	if err != nil {
		return transaction.Transaction{}, err
	}
	rawDesc := row[table.Column("Description")]
	desc, location := parseCitiDescription(rawDesc)
	debit := strings.TrimSpace(row[table.Column("Debit")])
	credit := strings.TrimSpace(row[table.Column("Credit")])

	// Debit = charges (positive), Credit = payments/refunds (always negative).
	// Citi uses positive credit values for payments ("1233.79") and
	// negative credit values for refunds ("-4.35").
	var amount transaction.Amount
	if debit != "" {
		amount, err = transaction.ParseAmount(debit)
	} else {
		amount, err = transaction.ParseAmount(credit)
		if err == nil && !strings.HasPrefix(credit, "-") {
			amount = amount.Negate()
		}
	}
	if err != nil {
		return transaction.Transaction{}, err
	}

	return transaction.Transaction{
		ID:          transaction.GenerateID(date, rawDesc, amount.String()),
		Date:        date,
		Description: desc,
		Location:    location,
		Amount:      amount,
	}, nil
}

// parseCitiDescription extracts the state code from the last 2 characters
// of Citi descriptions, except for payment-related entries.
func parseCitiDescription(desc string) (description, state string) {
	upper := strings.ToUpper(desc)
	if strings.Contains(upper, "AUTOPAY") || strings.Contains(upper, "PAYMENT") {
		return desc, ""
	}
	if len(desc) < 3 {
		return desc, ""
	}
	return strings.TrimSpace(desc[:len(desc)-2]), desc[len(desc)-2:]
}
