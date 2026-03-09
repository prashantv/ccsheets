# ccsheets

CLI tool to parse credit card CSV exports (Chase, Amex, Citi) into standardized transactions and upload them to Google Sheets.

## Usage

```
ccsheets <command> [flags] <files>...
```

### Commands

#### `parse` — Parse CSV files and print transactions

```
ccsheets parse [flags] <csv-file>...

Flags:
  -provider    Provider name: chase, amex, citi (auto-detected from filename if omitted)
  -output      Output format: table, json (default: table)
```

#### `upload` — Parse and upload transactions to Google Sheets

```
ccsheets upload [flags] <csv-file>...

Flags:
  -provider        Provider name: chase, amex, citi (auto-detected from filename if omitted)
  -spreadsheet-id  Google Sheets spreadsheet ID (env: CCSHEETS_SPREADSHEET_ID)
```

Upload deduplicates by transaction ID — rows already in the sheet are skipped.

### Provider auto-detection

The provider is auto-detected from the filename prefix:
- `Chase*` → chase
- `Amex*` → amex
- `Citi*` → citi

Account identifiers are also extracted from filenames:
- `Chase1111_Activity...` → Provider: Chase, Account: 1111
- `Amex2222-2024.csv` → Provider: Amex, Account: 2222
- `Citi-2025.CSV` → Provider: Citi

### Examples

```sh
# Parse a single file
ccsheets parse data/Chase1111_2026.CSV

# Parse multiple files as JSON
ccsheets parse -output json data/Chase1111-2025.csv data/Chase1111-2026.csv

# Upload to Google Sheets
export CCSHEETS_SPREADSHEET_ID=your-spreadsheet-id
ccsheets upload data/*.CSV data/*.csv
```

## Transaction schema

Each transaction has the following fields:

| Field       | Description                                    | Sheet column |
|-------------|------------------------------------------------|--------------|
| ID          | Unique identifier (hash-based or provider ID)  | A            |
| Provider    | Card provider: Chase, Amex, Citi               | B            |
| Account     | Account nickname or last 4 digits              | C            |
| Date        | Transaction date (YYYY-MM-DD)                  | D            |
| Description | Merchant / transaction description             | E            |
| Location    | City, State (when available)                   | F            |
| Amount      | Positive = charge, negative = payment/credit   | G            |
| Category    | Provider-assigned category (if available)      | H            |

Columns I+ are reserved for manual notes in the sheet.

## Providers

### Chase
- CSV format: `Transaction Date,Post Date,Description,Category,Type,Amount,Memo`
- IDs are generated from a hash of date + description + amount
- No location extraction

### Amex (Platinum and Blue)
- CSV format: `Date,Description,Amount,...Reference,Category`
- Uses the `Reference` field as a natural unique ID
- Location extracted from `City/State` column
- Description cleaned using fixed-width field split (42-char field, first 20 = merchant name)

### Citi
- CSV format: `Status,Date,Description,Debit,Credit,Member Name`
- Last 2 characters of description extracted as state code
- No category provided

## Google Sheets setup

Authentication uses Application Default Credentials. Set up a service account:

1. Create a service account in Google Cloud Console
2. Download the key JSON file
3. Set `GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json`
4. Share the spreadsheet with the service account email

## Project structure

```
main.go              CLI entry point, subcommands, file loading, output formatting
csvtable             CSV parsing with column remapping
transaction/         Transaction model
provider/            CSV parsers for different providers
sheet                Google Sheets client with dedup upload
```
