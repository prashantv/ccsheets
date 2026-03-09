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
	date := row[table.Column("Date")]
	desc := row[table.Column("Description")]
	debit := strings.TrimSpace(row[table.Column("Debit")])
	credit := strings.TrimSpace(row[table.Column("Credit")])

	// Debit = charges (positive), Credit = payments (show as negative).
	var amount transaction.Amount
	var err error
	if debit != "" {
		amount, err = transaction.ParseAmount(debit)
	} else {
		amount, err = transaction.ParseAmount(credit)
		if err == nil {
			amount = amount.Negate()
		}
	}
	if err != nil {
		return transaction.Transaction{}, err
	}

	return transaction.Transaction{
		ID:          transaction.GenerateID(date, desc, amount.String()),
		Date:        date,
		Description: desc,
		Amount:      amount,
	}, nil
}
