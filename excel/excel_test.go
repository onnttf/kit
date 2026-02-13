package excel

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// Helper function to create a test Excel file in memory
func createTestExcelFile(t *testing.T) *bytes.Buffer {
	t.Helper()

	f := excelize.NewFile()
	defer f.Close()

	// Create Sheet1 with test data
	sheet1 := "Sheet1"
	data1 := [][]any{
		{"Name", "Age", "City"},
		{"Alice", 30, "New York"},
		{"Bob", 25, "Los Angeles"},
		{"Charlie", 35, "Chicago"},
	}

	for rowIdx, row := range data1 {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue(sheet1, cell, value)
		}
	}

	// Create Sheet2 with test data
	sheet2 := "Sheet2"
	f.NewSheet(sheet2)
	data2 := [][]any{
		{"Product", "Price"},
		{"Apple", 1.5},
		{"Banana", 0.8},
	}

	for rowIdx, row := range data2 {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue(sheet2, cell, value)
		}
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("Failed to create test Excel file: %v", err)
	}

	return &buf
}

// Test Suite for IsExcel Function

func TestIsExcel(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"valid xlsx", "test.xlsx", true},
		{"valid XLSX uppercase", "TEST.XLSX", true},
		{"valid mixed case", "Test.XlSx", true},
		{"xls file", "test.xls", false},
		{"csv file", "test.csv", false},
		{"txt file", "test.txt", false},
		{"no extension", "test", false},
		{"xlsx in name but wrong ext", "xlsx.txt", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExcel(tt.filename)
			if got != tt.want {
				t.Errorf("IsExcel(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

// Test Suite for Open Function

func TestOpen_Success(t *testing.T) {
	buf := createTestExcelFile(t)

	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	if wb == nil {
		t.Fatal("Expected non-nil Workbook")
	}
	if wb.file == nil {
		t.Fatal("Expected non-nil underlying file")
	}
}

func TestOpen_InvalidFile(t *testing.T) {
	invalidData := bytes.NewReader([]byte("not an excel file"))

	wb, err := Open(invalidData)
	if err == nil {
		if wb != nil {
			wb.Close()
		}
		t.Fatal("Expected error for invalid Excel file")
	}

	if !strings.Contains(err.Error(), "open Excel file") {
		t.Errorf("Expected 'open Excel file' error, got: %v", err)
	}
}

// Test Suite for Workbook.GetSheet Function

func TestWorkbook_GetSheet(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	t.Run("read Sheet1", func(t *testing.T) {
		rows, err := wb.GetSheet("Sheet1")
		if err != nil {
			t.Fatalf("GetSheet() error = %v", err)
		}

		if len(rows) != 4 {
			t.Errorf("Expected 4 rows, got %d", len(rows))
		}

		// Check header row
		if len(rows) > 0 {
			expected := []string{"Name", "Age", "City"}
			if !reflect.DeepEqual(rows[0], expected) {
				t.Errorf("Header row: got %v, want %v", rows[0], expected)
			}
		}

		// Check data row
		if len(rows) > 1 {
			expected := []string{"Alice", "30", "New York"}
			if !reflect.DeepEqual(rows[1], expected) {
				t.Errorf("Data row: got %v, want %v", rows[1], expected)
			}
		}
	})

	t.Run("read non-existent sheet", func(t *testing.T) {
		_, err := wb.GetSheet("NonExistent")
		if err == nil {
			t.Error("Expected error for non-existent sheet")
		}
	})
}

// Test Suite for Workbook.GetSheets Function

func TestWorkbook_GetSheets_AllSheets(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	sheets, err := wb.GetSheets()
	if err != nil {
		t.Fatalf("GetSheets() error = %v", err)
	}

	if len(sheets) != 2 {
		t.Errorf("Expected 2 sheets, got %d", len(sheets))
	}

	if _, ok := sheets["Sheet1"]; !ok {
		t.Error("Expected Sheet1 to be present")
	}

	if _, ok := sheets["Sheet2"]; !ok {
		t.Error("Expected Sheet2 to be present")
	}
}

func TestWorkbook_GetSheets_SpecificSheets(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	sheets, err := wb.GetSheets("Sheet1")
	if err != nil {
		t.Fatalf("GetSheets() error = %v", err)
	}

	if len(sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(sheets))
	}

	if _, ok := sheets["Sheet1"]; !ok {
		t.Error("Expected Sheet1 to be present")
	}

	if _, ok := sheets["Sheet2"]; ok {
		t.Error("Did not expect Sheet2 to be present")
	}
}

// Test Suite for Workbook.GetSheetList Function

func TestWorkbook_GetSheetList(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	sheetList := wb.GetSheetList()

	if len(sheetList) != 2 {
		t.Errorf("Expected 2 sheets, got %d", len(sheetList))
	}

	expectedSheets := map[string]bool{"Sheet1": true, "Sheet2": true}
	for _, sheet := range sheetList {
		if !expectedSheets[sheet] {
			t.Errorf("Unexpected sheet: %s", sheet)
		}
	}
}

// Test Suite for Workbook.StreamRows Function

func TestWorkbook_StreamRows(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	var rowCount int
	var firstRow []string

	err = wb.StreamRows("Sheet1", func(rowIndex int, row []string) error {
		rowCount++
		if rowIndex == 0 {
			firstRow = row
		}
		return nil
	})

	if err != nil {
		t.Fatalf("StreamRows() error = %v", err)
	}

	if rowCount != 4 {
		t.Errorf("Expected 4 rows processed, got %d", rowCount)
	}

	expectedFirstRow := []string{"Name", "Age", "City"}
	if !reflect.DeepEqual(firstRow, expectedFirstRow) {
		t.Errorf("First row: got %v, want %v", firstRow, expectedFirstRow)
	}
}

func TestWorkbook_StreamRows_EarlyStop(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	var rowCount int
	stopErr := bytes.ErrTooLarge // Use a sentinel error

	err = wb.StreamRows("Sheet1", func(rowIndex int, row []string) error {
		rowCount++
		if rowIndex >= 1 {
			return stopErr
		}
		return nil
	})

	if err == nil {
		t.Fatal("Expected error from handler")
	}

	if rowCount != 2 {
		t.Errorf("Expected 2 rows before stop, got %d", rowCount)
	}
}

// Test Suite for ReadExcel Function (Backward Compatibility)

func TestReadExcel_AllSheets(t *testing.T) {
	buf := createTestExcelFile(t)

	sheets, err := ReadExcel(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("ReadExcel() error = %v", err)
	}

	if len(sheets) != 2 {
		t.Errorf("Expected 2 sheets, got %d", len(sheets))
	}

	if _, ok := sheets["Sheet1"]; !ok {
		t.Error("Expected Sheet1")
	}
	if _, ok := sheets["Sheet2"]; !ok {
		t.Error("Expected Sheet2")
	}
}

func TestReadExcel_SpecificSheets(t *testing.T) {
	buf := createTestExcelFile(t)

	sheets, err := ReadExcel(bytes.NewReader(buf.Bytes()), "Sheet1")
	if err != nil {
		t.Fatalf("ReadExcel() error = %v", err)
	}

	if len(sheets) != 1 {
		t.Errorf("Expected 1 sheet, got %d", len(sheets))
	}

	if _, ok := sheets["Sheet1"]; !ok {
		t.Error("Expected Sheet1")
	}
}

// Test Suite for ReadExcelSheet Function (Backward Compatibility)

func TestReadExcelSheet(t *testing.T) {
	buf := createTestExcelFile(t)

	rows, err := ReadExcelSheet(bytes.NewReader(buf.Bytes()), "Sheet1")
	if err != nil {
		t.Fatalf("ReadExcelSheet() error = %v", err)
	}

	if len(rows) != 4 {
		t.Errorf("Expected 4 rows, got %d", len(rows))
	}

	expectedHeader := []string{"Name", "Age", "City"}
	if !reflect.DeepEqual(rows[0], expectedHeader) {
		t.Errorf("Header: got %v, want %v", rows[0], expectedHeader)
	}
}

// Test Suite for StreamRows Function (Standalone)

func TestStreamRows_Standalone(t *testing.T) {
	buf := createTestExcelFile(t)

	var rowCount int
	err := StreamRows(bytes.NewReader(buf.Bytes()), "Sheet1", func(rowIndex int, row []string) error {
		rowCount++
		return nil
	})

	if err != nil {
		t.Fatalf("StreamRows() error = %v", err)
	}

	if rowCount != 4 {
		t.Errorf("Expected 4 rows, got %d", rowCount)
	}
}

// Test Suite for Workbook.Close Function

func TestWorkbook_Close(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	err = wb.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Multiple closes should be safe
	err = wb.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

// Integration Tests

func TestWorkbook_ReuseForMultipleReads(t *testing.T) {
	buf := createTestExcelFile(t)
	wb, err := Open(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer wb.Close()

	// Read Sheet1 multiple times
	for i := range 3 {
		rows, err := wb.GetSheet("Sheet1")
		if err != nil {
			t.Fatalf("GetSheet() iteration %d error = %v", i, err)
		}
		if len(rows) != 4 {
			t.Errorf("Iteration %d: expected 4 rows, got %d", i, len(rows))
		}
	}

	// Read Sheet2
	rows, err := wb.GetSheet("Sheet2")
	if err != nil {
		t.Fatalf("GetSheet(Sheet2) error = %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("Sheet2: expected 3 rows, got %d", len(rows))
	}
}

// Benchmark Tests

func BenchmarkOpen(b *testing.B) {
	buf := createTestExcelFile(&testing.T{})
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wb, _ := Open(bytes.NewReader(data))
		if wb != nil {
			wb.Close()
		}
	}
}

func BenchmarkWorkbook_GetSheet(b *testing.B) {
	buf := createTestExcelFile(&testing.T{})
	wb, _ := Open(bytes.NewReader(buf.Bytes()))
	defer wb.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = wb.GetSheet("Sheet1")
	}
}

func BenchmarkWorkbook_StreamRows(b *testing.B) {
	buf := createTestExcelFile(&testing.T{})
	wb, _ := Open(bytes.NewReader(buf.Bytes()))
	defer wb.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wb.StreamRows("Sheet1", func(rowIndex int, row []string) error {
			return nil
		})
	}
}
