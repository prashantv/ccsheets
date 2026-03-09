package provider

import (
	"fmt"
	"io"
	"strings"

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
	amount := row[table.Column("Amount")]
	category := row[table.Column("Category")]

	// Chase uses negative amounts for charges; negate so charges are positive.
	amount = negateAmount(amount)

	return transaction.Transaction{
		ID:          transaction.GenerateID(date, desc, amount),
		Date:        date,
		Description: desc,
		Amount:      amount,
		Category:    category,
	}, nil
}

// negateAmount flips the sign of a numeric string.
func negateAmount(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || s == "0.00" {
		return s
	}
	if strings.HasPrefix(s, "-") {
		return s[1:]
	}
	return fmt.Sprintf("-%s", s)
}
