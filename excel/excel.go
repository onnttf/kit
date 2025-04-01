package excel

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// IsExcel checks if a file is an Excel file based on its file name extension.
// It returns true if the file name ends with ".xlsx" (case-insensitive), otherwise false.
func IsExcel(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".xlsx")
}

// ReadExcel reads data from an Excel file.
// It takes an io.Reader representing the Excel file and a variable number of sheet names.
// If no sheet names are provided, it reads data from all sheets in the file.
// It returns a map where the keys are sheet names and the values are 2D string slices representing the sheet data,
// or an error if the file cannot be opened or read.
func ReadExcel(file io.Reader, sheetNames ...string) (map[string][][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer workbook.Close()

	if len(sheetNames) == 0 {
		sheetNames = workbook.GetSheetList()
	}

	sheetData := make(map[string][][]string)
	for _, sheetName := range sheetNames {
		data, err := readSheetData(workbook, sheetName)
		if err != nil {
			return nil, err
		}
		sheetData[sheetName] = data
	}

	return sheetData, nil
}

// ReadExcelSheet reads data from a specific sheet in an Excel file.
// It takes an io.Reader representing the Excel file and the name of the sheet to read.
// It returns a 2D string slice representing the sheet data,
// or an error if the file cannot be opened or the sheet cannot be read.
func ReadExcelSheet(file io.Reader, sheetName string) ([][]string, error) {
	workbook, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer workbook.Close()

	return readSheetData(workbook, sheetName)
}

// readSheetData reads data from a specific sheet in an Excel workbook.
// It takes an excelize.File pointer and the sheet name.
// It retrieves all rows from the sheet, filters out empty rows, and returns the valid rows as a 2D string slice.
// It returns an error if the rows cannot be read.
func readSheetData(workbook *excelize.File, sheetName string) ([][]string, error) {
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from sheet %s: %w", sheetName, err)
	}

	var validRows [][]string
	for _, row := range rows {
		if isRowEmpty(row) {
			continue
		}
		validRows = append(validRows, row)
	}
	return validRows, nil
}

// isRowEmpty checks if a row in an Excel sheet is empty.
// It iterates through the cells in the row and returns true if all cells are empty or contain only whitespace, otherwise false.
func isRowEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}
