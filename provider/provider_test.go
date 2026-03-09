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
					Date:        "2026-01-15",
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
					Date:        "2026-01-10",
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
			name: "platinum with location in description",
			giveCSV: csvLines(
				"Date,Description,Card Member,Account #,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`02/01/2026,HBO Max             NEW YORK            NY,JANE DOE,-99001,5.75,"details",HBO Max,123 MAIN ST,"NEW YORK`+"\n"+`NY",10001,UNITED STATES,'320260101234567890',Merchandise & Supplies-Internet Purchase`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260101234567890",
					Date:        "2026-02-01",
					Description: "HBO Max",
					Location:    "NEW YORK, NY",
					Amount:      mustAmount(t, "5.75"),
					Category:    "Merchandise & Supplies-Internet Purchase",
				},
			},
		},
		{
			name: "platinum credit with no location",
			giveCSV: csvLines(
				"Date,Description,Card Member,Account #,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`02/01/2026,Platinum Digital Entertainment Credit,JANE DOE,-99001,-5.75,"details",Platinum Digital Entertainment Credit,,,,,'320260101234567891',Fees & Adjustments`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260101234567891",
					Date:        "2026-02-01",
					Description: "Platinum Digital Entertainment Credit",
					Amount:      mustAmount(t, "-5.75"),
					Category:    "Fees & Adjustments",
				},
			},
		},
		{
			name: "blue with location in description",
			giveCSV: csvLines(
				"Date,Description,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`03/02/2026,AplPay SAFEWAY      SAN FRANCISCO       CA,36.17,"details",AplPay SAFEWAY,"298 KING ST","SAN FRANCISCO`+"\n"+`CA",94107,UNITED STATES,'320260620701362525',Merchandise & Supplies-Groceries`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260620701362525",
					Date:        "2026-03-02",
					Description: "AplPay SAFEWAY",
					Location:    "SAN FRANCISCO, CA",
					Amount:      mustAmount(t, "36.17"),
					Category:    "Merchandise & Supplies-Groceries",
				},
			},
		},
		{
			name: "short description not matching city",
			giveCSV: csvLines(
				"Date,Description,Amount,Extended Details,Appears On Your Statement As,Address,City/State,Zip Code,Country,Reference,Category",
				`03/02/2026,UBER,10.76,"details",Uber Trip,"1515 3RD ST","SAN FRANCISCO`+"\n"+`CA",94107,UNITED STATES,'320260620701362526',Transportation`,
			),
			wantTxns: []transaction.Transaction{
				{
					ID:          "320260620701362526",
					Date:        "2026-03-02",
					Description: "UBER",
					Location:    "SAN FRANCISCO, CA",
					Amount:      mustAmount(t, "10.76"),
					Category:    "Transportation",
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
			name: "debit charge with state extracted",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,02/15/2026,"COSTCO WHSE #0144 SAN FRANCISCOCA",45.00,,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "2026-02-15",
					Description: "COSTCO WHSE #0144 SAN FRANCISCO",
					Location:    "CA",
					Amount:      mustAmount(t, "45.00"),
				},
			},
		},
		{
			name: "debit with space before state",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,01/11/2026,"TST*AURUM Los Altos CA",75.42,,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "2026-01-11",
					Description: "TST*AURUM Los Altos",
					Location:    "CA",
					Amount:      mustAmount(t, "75.42"),
				},
			},
		},
		{
			name: "credit payment not split",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,02/10/2026,"ONLINE PAYMENT THANK YOU",,500.00,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "2026-02-10",
					Description: "ONLINE PAYMENT THANK YOU",
					Amount:      mustAmount(t, "-500.00"),
				},
			},
		},
		{
			name: "autopay not split",
			giveCSV: csvLines(
				"Status,Date,Description,Debit,Credit,Member Name",
				`Cleared,02/14/2026,"AUTOPAY 210525023533040RAUTOPAY AUTO-PMT",,1233.79,JANE DOE`,
			),
			wantTxns: []transaction.Transaction{
				{
					Date:        "2026-02-14",
					Description: "AUTOPAY 210525023533040RAUTOPAY AUTO-PMT",
					Amount:      mustAmount(t, "-1233.79"),
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
	if got.Location != want.Location {
		t.Errorf("Location: got %q, want %q", got.Location, want.Location)
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
