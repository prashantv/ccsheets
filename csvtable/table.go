package csvtable

import (
	"encoding/csv"
	"fmt"
	"io"
)

// Table is a parsed CSV with remapped column headers.
type Table struct {
	Header []string
	Rows   [][]string
}

// ColumnMap defines how CSV columns map to output columns.
// Keys are the desired output column names, values are the original CSV header names.
type ColumnMap map[string]string

// Parse reads a CSV and remaps columns according to the given ColumnMap.
// Only columns present in the map are included in the output Table.
// Output columns are ordered by the order of outputColumns.
func Parse(r io.Reader, colMap ColumnMap, outputColumns []string) (Table, error) {
	cr := csv.NewReader(r)

	header, err := cr.Read()
	if err != nil {
		return Table{}, fmt.Errorf("reading CSV header: %w", err)
	}

	// Build index: CSV column name -> position.
	csvIndex := make(map[string]int, len(header))
	for i, name := range header {
		csvIndex[name] = i
	}

	// Resolve which CSV column index feeds each output column.
	srcIndices := make([]int, len(outputColumns))
	for i, outCol := range outputColumns {
		csvCol, ok := colMap[outCol]
		if !ok {
			return Table{}, fmt.Errorf("output column %q not found in column map", outCol)
		}
		idx, ok := csvIndex[csvCol]
		if !ok {
			return Table{}, fmt.Errorf("CSV column %q (for output %q) not found in header", csvCol, outCol)
		}
		srcIndices[i] = idx
	}

	var rows [][]string
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Table{}, fmt.Errorf("reading CSV row: %w", err)
		}

		row := make([]string, len(outputColumns))
		for i, srcIdx := range srcIndices {
			row[i] = record[srcIdx]
		}
		rows = append(rows, row)
	}

	return Table{
		Header: outputColumns,
		Rows:   rows,
	}, nil
}

// Column returns the index of the named column, or -1 if not found.
func (t Table) Column(name string) int {
	for i, h := range t.Header {
		if h == name {
			return i
		}
	}
	return -1
}
