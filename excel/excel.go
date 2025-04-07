package excel

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// IsExcel checks if a file is an Excel file based on its extension.
// Returns true for ".xlsx" files (case-insensitive).
func IsExcel(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".xlsx")
}

// ReadExcel reads data from an Excel file.
// It reads from specified sheets or all sheets if none specified.
// Returns a map of sheet names to their data as 2D string slices.
func ReadExcel(file io.Reader, sheetNames ...string) (map[string][][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer workbook.Close()

	// If no sheets specified, read all sheets
	if len(sheetNames) == 0 {
		sheetNames = workbook.GetSheetList()
	}

	// Process each requested sheet
	sheetData := make(map[string][][]string, len(sheetNames))
	for _, sheetName := range sheetNames {
		data, err := extractSheetData(workbook, sheetName)
		if err != nil {
			return nil, err
		}
		sheetData[sheetName] = data
	}

	return sheetData, nil
}

// ReadExcelSheet reads data from a single Excel sheet.
// Returns the sheet data as a 2D string slice.
func ReadExcelSheet(file io.Reader, sheetName string) ([][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer workbook.Close()

	return extractSheetData(workbook, sheetName)
}

// extractSheetData retrieves and cleans data from a specific sheet.
// Returns non-empty rows as a 2D string slice.
func extractSheetData(workbook *excelize.File, sheetName string) ([][]string, error) {
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", sheetName, err)
	}

	// Filter out empty rows
	var nonEmptyRows [][]string
	for _, row := range rows {
		if !isRowEmpty(row) {
			nonEmptyRows = append(nonEmptyRows, row)
		}
	}

	return nonEmptyRows, nil
}

// isRowEmpty checks if a row contains only empty cells.
// Returns true if all cells are empty or whitespace-only.
func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
