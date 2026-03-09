package transaction

import (
	"crypto/sha256"
	"fmt"

	"github.com/prashantv/ccsheets/csvtable"
)

// Transaction is a standardized credit card transaction.
type Transaction struct {
	ID          string `json:"id"`
	Date        string `json:"date"`
	Description string `json:"description"`
	Amount      string `json:"amount"` // kept as string to avoid float precision issues
	Category    string `json:"category"`
}

// Parser converts a table row into a Transaction.
// Each provider (Chase, Amex, Citi) implements this differently.
type Parser interface {
	Parse(table csvtable.Table, row []string) (Transaction, error)
}

// ParseAll converts every row in the table into Transactions using the given parser.
func ParseAll(table csvtable.Table, p Parser) ([]Transaction, error) {
	txns := make([]Transaction, 0, len(table.Rows))
	for i, row := range table.Rows {
		txn, err := p.Parse(table, row)
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", i, err)
		}
		txns = append(txns, txn)
	}
	return txns, nil
}

// GenerateID produces a deterministic ID from the given fields.
func GenerateID(parts ...string) string {
	h := sha256.New()
	for _, p := range parts {
		h.Write([]byte(p))
		h.Write([]byte{0}) // separator
	}
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}
