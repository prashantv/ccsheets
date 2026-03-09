package provider

import (
	"strings"
	"testing"

	"github.com/prashantv/ccsheets/transaction"
)

func mustAmount(t *testing.T, s string) transaction.Amount {
	t.Helper()
	a, err := transaction.ParseAmount(s)
	if err != nil {
		t.Fatalf("ParseAmount(%q): %v", s, err)
	}
	return a
}

func TestChaseParser(t *testing.T) {
	tests := []struct {
		name     string
		giveCSV  string
		wantTxns []transaction.Transaction
	}{
		{
			name: "charge negated to positive",
			giveCSV: csvLines(
				"Transaction Date,Post Date,Description,Category,Type,Amount,Memo",
				"01/15/2026,01/16/2026,TACO PALACE,Food & Drink,Sale,-42.50,",
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "01/15/2026",
					Description: "TACO PALACE",
					Amount:      mustAmount(t, "42.50"),
					Category:    "Food & Drink",
				},
			},
		},
		{
			name: "payment negated to negative",
			giveCSV: csvLines(
				"Transaction Date,Post Date,Description,Category,Type,Amount,Memo",
				"01/10/2026,01/10/2026,AUTOMATIC PAYMENT - THANK,,Payment,200.00,",
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "01/10/2026",
					Description: "AUTOMATIC PAYMENT - THANK",
					Amount:      mustAmount(t, "-200.00"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := LoadChase(strings.NewReader(tt.giveCSV))
			if err != nil {
				t.Fatalf("LoadChase: %v", err)
			}

			txns, err := transaction.ParseAll(table, ChaseParser{})
			if err != nil {
				t.Fatalf("ParseAll: %v", err)
			}

			if got, want := len(txns), len(tt.wantTxns); got != want {
				t.Fatalf("got %d transactions, want %d", got, want)
			}

			for i, want := range tt.wantTxns {
				assertTxn(t, txns[i], want)
				if txns[i].ID == "" {
					t.Error("expected non-empty ID")
				}
			}
		})
	}
}

func TestChaseParser_DeterministicIDs(t *testing.T) {
	csv := csvLines(
		"Transaction Date,Post Date,Description,Category,Type,Amount,Memo",
		"01/15/2026,01/16/2026,TACO PALACE,Food & Drink,Sale,-42.50,",
		"01/10/2026,01/10/2026,AUTOMATIC PAYMENT - THANK,,Payment,200.00,",
	)

	table, err := LoadChase(strings.NewReader(csv))
	if err != nil {
		t.Fatalf("LoadChase: %v", err)
	}

	txns, err := transaction.ParseAll(table, ChaseParser{})
	if err != nil {
		t.Fatalf("ParseAll: %v", err)
	}

	if txns[0].ID == txns[1].ID {
		t.Error("different transactions should have different IDs")
	}
}

func TestAmexParser(t *testing.T) {
	tests := []struct {
		name     string
		giveCSV  string
		wantTxns []transaction.Transaction
	}{
		{
			name: "platinum format",
			giveCSV: csvLines(
				"Date,Description,Card Member,Account #,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`02/01/2026,COFFEE ROASTERS,JANE DOE,-99001,5.75,"details",COFFEE ROASTERS,123 MAIN ST,"PORTLAND`+"\n"+`OR",97201,UNITED STATES,'320260101234567890',Merchandise & Supplies-Groceries`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260101234567890",
					Date:        "02/01/2026",
					Description: "COFFEE ROASTERS",
					Amount:      mustAmount(t, "5.75"),
					Category:    "Merchandise & Supplies-Groceries",
				},
			},
		},
		{
			name: "blue format",
			giveCSV: csvLines(
				"Date,Description,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`03/02/2026,GROCERY MART,36.17,"details",GROCERY MART,"298 KING ST","SAN FRANCISCO`+"\n"+`CA",94107,UNITED STATES,'320260620701362525',Merchandise & Supplies-Groceries`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260620701362525",
					Date:        "03/02/2026",
					Description: "GROCERY MART",
					Amount:      mustAmount(t, "36.17"),
					Category:    "Merchandise & Supplies-Groceries",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := LoadAmex(strings.NewReader(tt.giveCSV))
			if err != nil {
				t.Fatalf("LoadAmex: %v", err)
			}

			txns, err := transaction.ParseAll(table, AmexParser{})
			if err != nil {
				t.Fatalf("ParseAll: %v", err)
			}

			if got, want := len(txns), len(tt.wantTxns); got != want {
				t.Fatalf("got %d transactions, want %d", got, want)
			}

			for i, want := range tt.wantTxns {
				assertTxn(t, txns[i], want)
			}
		})
	}
}

func TestCitiParser(t *testing.T) {
	tests := []struct {
		name     string
		giveCSV  string
		wantTxns []transaction.Transaction
	}{
		{
			name: "debit charge stays positive",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,02/15/2026,"GAS STATION 1234",45.00,,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "02/15/2026",
					Description: "GAS STATION 1234",
					Amount:      mustAmount(t, "45.00"),
				},
			},
		},
		{
			name: "credit payment becomes negative",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,02/10/2026,"ONLINE PAYMENT THANK YOU",,500.00,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "02/10/2026",
					Description: "ONLINE PAYMENT THANK YOU",
					Amount:      mustAmount(t, "-500.00"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			table, err := LoadCiti(strings.NewReader(tt.giveCSV))
			if err != nil {
				t.Fatalf("LoadCiti: %v", err)
			}

			txns, err := transaction.ParseAll(table, CitiParser{})
			if err != nil {
				t.Fatalf("ParseAll: %v", err)
			}

			if got, want := len(txns), len(tt.wantTxns); got != want {
				t.Fatalf("got %d transactions, want %d", got, want)
			}

			for i, want := range tt.wantTxns {
				assertTxn(t, txns[i], want)
				if txns[i].ID == "" {
					t.Error("expected non-empty ID")
				}
			}
		})
	}
}

func assertTxn(t *testing.T, got, want transaction.Transaction) {
	t.Helper()
	if want.ID != "" && got.ID != want.ID {
		t.Errorf("ID: got %q, want %q", got.ID, want.ID)
	}
	if got.Date != want.Date {
		t.Errorf("Date: got %q, want %q", got.Date, want.Date)
	}
	if got.Description != want.Description {
		t.Errorf("Description: got %q, want %q", got.Description, want.Description)
	}
	if got.Amount != want.Amount {
		t.Errorf("Amount: got %s, want %s", got.Amount, want.Amount)
	}
	if got.Category != want.Category {
		t.Errorf("Category: got %q, want %q", got.Category, want.Category)
	}
}

func csvLines(lines ...string) string {
	return strings.Join(lines, "\n")
}
