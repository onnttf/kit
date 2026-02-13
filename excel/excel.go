package excel

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Workbook wraps excelize.File to support multiple sheet reads without reopening the file.
type Workbook struct {
	file *excelize.File
}

// Open creates a new Workbook from an io.Reader, allowing multiple operations on the same opened file.
//
// Example:
//
//	wb, err := excel.Open(file)
//	if err != nil {
//	    return err
//	}
//	defer wb.Close()
//
//	sheet1Data, err := wb.GetSheet("Sheet1")
//	sheet2Data, err := wb.GetSheet("Sheet2")
func Open(reader io.Reader) (*Workbook, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("open Excel file: %w", err)
	}
	return &Workbook{file: f}, nil
}

// GetSheet returns row data from a specific sheet as a 2D string slice.
func (wb *Workbook) GetSheet(sheetName string) ([][]string, error) {
	return extractSheetData(wb.file, sheetName)
}

// GetSheets returns row data from multiple sheets, or all sheets if none specified.
func (wb *Workbook) GetSheets(sheetNames ...string) (map[string][][]string, error) {
	// If no sheets specified, read all sheets
	if len(sheetNames) == 0 {
		sheetNames = wb.file.GetSheetList()
	}

	sheetData := make(map[string][][]string, len(sheetNames))
	for _, sheetName := range sheetNames {
		data, err := extractSheetData(wb.file, sheetName)
		if err != nil {
			return nil, err
		}
		sheetData[sheetName] = data
	}

	return sheetData, nil
}

// GetSheetList returns the names of all sheets in the workbook.
func (wb *Workbook) GetSheetList() []string {
	return wb.file.GetSheetList()
}

// StreamRows reads a sheet row by row, calling the handler function for each row.
// This is memory-efficient for large files as it doesn't load all data at once.
//
// Example:
//
//	err := wb.StreamRows("Sheet1", func(rowIndex int, row []string) error {
//	    fmt.Printf("Row %d: %v\n", rowIndex, row)
//	    return nil
//	})
func (wb *Workbook) StreamRows(sheetName string, handler func(rowIndex int, row []string) error) error {
	rows, err := wb.file.Rows(sheetName)
	if err != nil {
		return fmt.Errorf("get rows iterator for sheet %s: %w", sheetName, err)
	}
	defer rows.Close()

	rowIndex := 0
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("read row %d in sheet %s: %w", rowIndex, sheetName, err)
		}

		if err := handler(rowIndex, row); err != nil {
			return fmt.Errorf("handler error at row %d: %w", rowIndex, err)
		}
		rowIndex++
	}

	return rows.Error()
}

// Close closes the underlying Excel file and releases resources.
func (wb *Workbook) Close() error {
	if wb.file != nil {
		return wb.file.Close()
	}
	return nil
}

// IsExcel returns true if the file has a case-insensitive ".xlsx" extension
func IsExcel(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".xlsx")
}

// ReadExcel returns row data from specified Excel sheets or all sheets if none specified, as a map of sheet names to 2D string slices.
// This is a convenience function that opens, reads, and closes the file in one call.
// For reading multiple sheets from the same file, consider using Open() and Workbook methods instead to avoid reopening the file.
func ReadExcel(file io.Reader, sheetNames ...string) (map[string][][]string, error) {
	wb, err := Open(file)
	if err != nil {
		return nil, err
	}
	defer wb.Close()

	return wb.GetSheets(sheetNames...)
}

// ReadExcelSheet returns row data from a single Excel sheet as a 2D string slice.
// This is a convenience function that opens, reads, and closes the file in one call.
// For reading multiple sheets from the same file, consider using Open() and Workbook methods instead to avoid reopening the file.
func ReadExcelSheet(file io.Reader, sheetName string) ([][]string, error) {
	wb, err := Open(file)
	if err != nil {
		return nil, err
	}
	defer wb.Close()

	return wb.GetSheet(sheetName)
}

// StreamRows reads an Excel sheet row by row, calling the handler function for each row.
// This is memory-efficient for large files as it doesn't load all data at once.
//
// Example:
//
//	err := excel.StreamRows(file, "Sheet1", func(rowIndex int, row []string) error {
//	    fmt.Printf("Row %d: %v\n", rowIndex, row)
//	    if rowIndex > 1000 {
//	        return fmt.Errorf("stop after 1000 rows")
//	    }
//	    return nil
//	})
func StreamRows(reader io.Reader, sheetName string, handler func(rowIndex int, row []string) error) error {
	wb, err := Open(reader)
	if err != nil {
		return err
	}
	defer wb.Close()

	return wb.StreamRows(sheetName, handler)
}

// extractSheetData retrieves row data from a specific Excel sheet as a 2D string slice
func extractSheetData(workbook *excelize.File, sheetName string) ([][]string, error) {
	rows, err := workbook.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("read rows from sheet %s: %w", sheetName, err)
	}
	return rows, nil
}
