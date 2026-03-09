package sheet

import (
	"context"
	"fmt"

	"github.com/prashantv/ccsheets/transaction"
)

// Client wraps the Google Sheets API for uploading transactions.
type Client struct {
	// TODO: add Google Sheets API client, spreadsheet ID, sheet name.
}

// Upload appends transactions to the sheet, skipping any whose ID already exists.
func (c *Client) Upload(ctx context.Context, txns []transaction.Transaction) error {
	existing, err := c.existingIDs(ctx)
	if err != nil {
		return fmt.Errorf("fetching existing IDs: %w", err)
	}

	var newTxns []transaction.Transaction
	for _, txn := range txns {
		if !existing[txn.ID] {
			newTxns = append(newTxns, txn)
		}
	}

	if len(newTxns) == 0 {
		return nil
	}

	return c.appendRows(ctx, newTxns)
}

func (c *Client) existingIDs(ctx context.Context) (map[string]bool, error) {
	// TODO: read the ID column from the sheet and return the set of known IDs.
	return nil, fmt.Errorf("not implemented")
}

func (c *Client) appendRows(ctx context.Context, txns []transaction.Transaction) error {
	// TODO: append rows to the sheet via the Google Sheets API.
	return fmt.Errorf("not implemented")
}
