package provider

import (
	"io"

	"github.com/prashantv/ccsheets/csvtable"
	"github.com/prashantv/ccsheets/transaction"
)

var _ transaction.Parser = ChaseParser{}

var _chaseColumns = csvtable.ColumnMap{
	"Date":        "Transaction Date",
	"PostDate":    "Post Date",
	"Description": "Description",
	"Category":    "Category",
	"Type":        "Type",
	"Amount":      "Amount",
}

var _chaseOutputColumns = []string{"Date", "PostDate", "Description", "Category", "Type", "Amount"}

func LoadChase(r io.Reader) (csvtable.Table, error) {
	return csvtable.Parse(r, _chaseColumns, _chaseOutputColumns)
}

type ChaseParser struct{}

func (ChaseParser) Parse(table csvtable.Table, row []string) (transaction.Transaction, error) {
	date := row[table.Column("Date")]
	desc := row[table.Column("Description")]
	category := row[table.Column("Category")]

	amount, err := transaction.ParseAmount(row[table.Column("Amount")])
	if err != nil {
		return transaction.Transaction{}, err
	}
	// Chase uses negative amounts for charges; negate so charges are positive.
	amount = amount.Negate()

	return transaction.Transaction{
		ID:          transaction.GenerateID(date, desc, amount.String()),
		Date:        date,
		Description: desc,
		Amount:      amount,
		Category:    category,
	}, nil
}
