package sheet

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/prashantv/ccsheets/transaction"
)

// Client wraps the Google Sheets API for uploading transactions.
type Client struct {
	srv           *sheets.Service
	spreadsheetID string
	sheetName     string
}

// NewClient creates a Sheets client using Application Default Credentials.
// Supports GOOGLE_APPLICATION_CREDENTIALS env var or gcloud ADC.
func NewClient(ctx context.Context, spreadsheetID, sheetName string, opts ...option.ClientOption) (*Client, error) {
	srv, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating sheets service: %w", err)
	}
	return &Client{
		srv:           srv,
		spreadsheetID: spreadsheetID,
		sheetName:     sheetName,
	}, nil
}

// Upload appends transactions to the sheet, skipping any whose ID already exists.
// Returns the number of new rows added.
func (c *Client) Upload(ctx context.Context, txns []transaction.Transaction) (int, error) {
	existing, err := c.existingIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("fetching existing IDs: %w", err)
	}

	var newTxns []transaction.Transaction
	for _, txn := range txns {
		if !existing[txn.ID] {
			newTxns = append(newTxns, txn)
		}
	}

	if len(newTxns) == 0 {
		return 0, nil
	}

	if err := c.appendRows(ctx, newTxns); err != nil {
		return 0, err
	}
	return len(newTxns), nil
}

func (c *Client) sheetRange(r string) string {
	return fmt.Sprintf("%s!%s", c.sheetName, r)
}

func (c *Client) existingIDs(ctx context.Context) (map[string]bool, error) {
	resp, err := c.srv.Spreadsheets.Values.
		Get(c.spreadsheetID, c.sheetRange("A:A")).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("reading ID column: %w", err)
	}

	ids := make(map[string]bool, len(resp.Values))
	for _, row := range resp.Values {
		if len(row) > 0 {
			if id, ok := row[0].(string); ok {
				ids[id] = true
			}
		}
	}
	return ids, nil
}

func (c *Client) appendRows(ctx context.Context, txns []transaction.Transaction) error {
	rows := make([][]interface{}, len(txns))
	for i, txn := range txns {
		rows[i] = []interface{}{
			txn.ID,
			txn.Date,
			txn.Description,
			txn.Amount.String(),
			txn.Category,
		}
	}

	_, err := c.srv.Spreadsheets.Values.
		Append(c.spreadsheetID, c.sheetRange("A:E"), &sheets.ValueRange{
			Values: rows,
		}).
		ValueInputOption("USER_ENTERED").
		InsertDataOption("INSERT_ROWS").
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("appending rows: %w", err)
	}
	return nil
}
