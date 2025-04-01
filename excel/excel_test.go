package excel

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestIsExcel(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		want     bool
	}{
		{"LowerCase", "file.xlsx", true},
		{"UpperCase", "FILE.XLSX", true},
		{"MixedCase", "FiLe.XlSx", true},
		{"NotExcel", "file.txt", false},
		{"NoExtension", "file", false},
		{"EmptyString", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsExcel(tt.fileName); got != tt.want {
				t.Errorf("IsExcel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// createTestExcelFile creates an Excel file with test data for testing purposes
func createTestExcelFile(t *testing.T) io.Reader {
	// Create a new Excel file
	f := excelize.NewFile()

	// Sheet1 data
	sheet1Data := [][]string{
		{"Header1", "Header2", "Header3"},
		{"Row1Col1", "Row1Col2", "Row1Col3"},
		{"Row2Col1", "Row2Col2", "Row2Col3"},
		{"", "", ""}, // Empty row that should be skipped
		{"Row3Col1", "Row3Col2", "Row3Col3"},
	}

	// Add data to Sheet1
	for r, row := range sheet1Data {
		for c, cellValue := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				t.Fatalf("Failed to convert coordinates to cell name: %v", err)
			}
			if err := f.SetCellValue("Sheet1", cell, cellValue); err != nil {
				t.Fatalf("Failed to set cell value: %v", err)
			}
		}
	}

	// Create a second sheet
	f.NewSheet("Sheet2")

	// Sheet2 data
	sheet2Data := [][]string{
		{"Col1", "Col2"},
		{"Data1", "Data2"},
		{"", "  "}, // Row with only whitespace
		{"Data3", "Data4"},
	}

	// Add data to Sheet2
	for r, row := range sheet2Data {
		for c, cellValue := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				t.Fatalf("Failed to convert coordinates to cell name: %v", err)
			}
			if err := f.SetCellValue("Sheet2", cell, cellValue); err != nil {
				t.Fatalf("Failed to set cell value: %v", err)
			}
		}
	}

	// Convert the file to a buffer
	buffer, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("Failed to write Excel file to buffer: %v", err)
	}

	return bytes.NewReader(buffer.Bytes())
}

func TestReadExcel(t *testing.T) {
	// Test with valid Excel file
	t.Run("ValidExcelFile", func(t *testing.T) {
		reader := createTestExcelFile(t)

		// Test reading all sheets
		result, err := ReadExcel(reader)
		if err != nil {
			t.Fatalf("ReadExcel() error = %v", err)
		}

		// Check the number of sheets
		if len(result) != 2 {
			t.Errorf("ReadExcel() returned %d sheets, want 2", len(result))
		}

		// Check Sheet1 data
		sheet1Data, exists := result["Sheet1"]
		if !exists {
			t.Errorf("ReadExcel() missing Sheet1 in result")
		} else {
			// Check rows in Sheet1
			if len(sheet1Data) != 4 { // 4 non-empty rows
				t.Errorf("Sheet1 has %d rows, want 4", len(sheet1Data))
			}

			// Check content of first row
			if sheet1Data[0][0] != "Header1" || sheet1Data[0][1] != "Header2" || sheet1Data[0][2] != "Header3" {
				t.Errorf("Sheet1 first row = %v, want [Header1 Header2 Header3]", sheet1Data[0])
			}
		}

		// Check Sheet2 data
		sheet2Data, exists := result["Sheet2"]
		if !exists {
			t.Errorf("ReadExcel() missing Sheet2 in result")
		} else {
			// Check rows in Sheet2
			if len(sheet2Data) != 3 { // 3 non-empty rows
				t.Errorf("Sheet2 has %d rows, want 3", len(sheet2Data))
			}
		}
	})

	// Test with specific sheet names
	t.Run("SpecificSheets", func(t *testing.T) {
		reader := createTestExcelFile(t)

		result, err := ReadExcel(reader, "Sheet1")
		if err != nil {
			t.Fatalf("ReadExcel() error = %v", err)
		}

		// Check that only Sheet1 was read
		if len(result) != 1 {
			t.Errorf("ReadExcel() returned %d sheets, want 1", len(result))
		}

		_, exists := result["Sheet1"]
		if !exists {
			t.Errorf("ReadExcel() missing Sheet1 in result")
		}

		_, exists = result["Sheet2"]
		if exists {
			t.Errorf("ReadExcel() should not include Sheet2 in result")
		}
	})

	// Test with invalid reader
	t.Run("InvalidReader", func(t *testing.T) {
		reader := strings.NewReader("This is not an Excel file")

		_, err := ReadExcel(reader)
		if err == nil {
			t.Errorf("ReadExcel() expected error for invalid Excel file, got nil")
		}
	})
}

func TestReadExcelSheet(t *testing.T) {
	// Test with valid Excel file and sheet
	t.Run("ValidSheet", func(t *testing.T) {
		reader := createTestExcelFile(t)

		sheet1Data, err := ReadExcelSheet(reader, "Sheet1")
		if err != nil {
			t.Fatalf("ReadExcelSheet() error = %v", err)
		}

		// Check rows in Sheet1
		if len(sheet1Data) != 4 { // 4 non-empty rows
			t.Errorf("Sheet1 has %d rows, want 4", len(sheet1Data))
		}

		// Check content of a specific row
		if sheet1Data[1][0] != "Row1Col1" || sheet1Data[1][1] != "Row1Col2" || sheet1Data[1][2] != "Row1Col3" {
			t.Errorf("Sheet1 second row = %v, want [Row1Col1 Row1Col2 Row1Col3]", sheet1Data[1])
		}
	})

	// Test with non-existent sheet
	t.Run("NonExistentSheet", func(t *testing.T) {
		reader := createTestExcelFile(t)

		_, err := ReadExcelSheet(reader, "NonExistentSheet")
		if err == nil {
			t.Errorf("ReadExcelSheet() expected error for non-existent sheet, got nil")
		}
	})

	// Test with invalid reader
	t.Run("InvalidReader", func(t *testing.T) {
		reader := strings.NewReader("This is not an Excel file")

		_, err := ReadExcelSheet(reader, "Sheet1")
		if err == nil {
			t.Errorf("ReadExcelSheet() expected error for invalid Excel file, got nil")
		}
	})
}

func TestIsRowEmpty(t *testing.T) {
	tests := []struct {
		name string
		row  []string
		want bool
	}{
		{"EmptyRow", []string{"", "", ""}, true},
		{"RowWithWhitespace", []string{"", "  ", "\t"}, true},
		{"NonEmptyRow", []string{"data", "", ""}, false},
		{"WhitespaceAndData", []string{"  ", "data", ""}, false},
		{"EmptySlice", []string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRowEmpty(tt.row); got != tt.want {
				t.Errorf("isRowEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
