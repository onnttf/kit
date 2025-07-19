package excel

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// IsExcel returns true if the file has a case-insensitive ".xlsx" extension
func IsExcel(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".xlsx")
}

// ReadExcel returns row data from specified Excel sheets or all sheets if none specified, as a map of sheet names to 2D string slices
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

// ReadExcelSheet returns row data from a single Excel sheet as a 2D string slice
func ReadExcelSheet(file io.Reader, sheetName string) ([][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer workbook.Close()

	return extractSheetData(workbook, sheetName)
}

// extractSheetData retrieves row data from a specific Excel sheet as a 2D string slice
func extractSheetData(workbook *excelize.File, sheetName string) ([][]string, error) {
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", sheetName, err)
	}
	return rows, nil
}
