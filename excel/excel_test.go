package excel

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestIsXLSX(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.xlsx", true},
		{"test.XLSX", true},
		{"test.xls", false},
		{"test.csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsXLSX(tt.filename))
		})
	}
}

func createTestExcelFile(t testing.TB) string {
	f := excelize.NewFile()
	defer f.Close()

	sheet, _ := f.NewSheet("Test")
	f.SetActiveSheet(sheet)

	_ = f.SetCellValue("Test", "A1", "Name")
	_ = f.SetCellValue("Test", "B1", "Age")
	_ = f.SetCellValue("Test", "A2", "Alice")
	_ = f.SetCellValue("Test", "B2", 25)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.xlsx")
	err := f.SaveAs(tmpFile)
	require.NoError(t, err)

	return tmpFile
}

type Person struct {
	Name string `excel:"A"`
	Age  int    `excel:"B"`
}

func TestParse(t *testing.T) {
	t.Run("valid row", func(t *testing.T) {
		row := []string{"Alice", "25"}
		result, err := Parse[Person](row)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 25, result.Age)
	})

	t.Run("short row", func(t *testing.T) {
		row := []string{"Alice"}
		result, err := Parse[Person](row)
		require.NoError(t, err)
		assert.Equal(t, "Alice", result.Name)
		assert.Equal(t, 0, result.Age)
	})
}

func TestWorkbook_Open(t *testing.T) {
	filePath := createTestExcelFile(t)
	wb, err := Open(filePath)
	require.NoError(t, err)
	require.NotNil(t, wb)
}

func TestWorkbook_Close(t *testing.T) {
	filePath := createTestExcelFile(t)
	wb, err := Open(filePath)
	require.NoError(t, err)
	err = wb.Close()
	assert.NoError(t, err)
}

func TestWorkbook_Sheets(t *testing.T) {
	filePath := createTestExcelFile(t)
	wb, err := Open(filePath)
	require.NoError(t, err)
	defer wb.Close()
	sheets := wb.Sheets()
	assert.Contains(t, sheets, "Test")
}

func TestWorkbook_Sheet(t *testing.T) {
	filePath := createTestExcelFile(t)
	wb, err := Open(filePath)
	require.NoError(t, err)
	defer wb.Close()
	sheet := wb.Sheet("Test")
	assert.NotNil(t, sheet)
}

func TestRow_Values(t *testing.T) {
	row := &Row{values: []string{"a", "b", "c"}}
	assert.Equal(t, []string{"a", "b", "c"}, row.Values())
}

func TestRow_Index(t *testing.T) {
	row := &Row{index: 5}
	assert.Equal(t, 5, row.Index())
}

func TestRow_Value(t *testing.T) {
	row := &Row{values: []string{"a", "b", "c"}}
	assert.Equal(t, "a", row.Value(0))
	assert.Equal(t, "b", row.Value(1))
	assert.Equal(t, "", row.Value(3))
}

func TestRow_Len(t *testing.T) {
	row := &Row{values: []string{"a", "b", "c"}}
	assert.Equal(t, 3, row.Len())
}

func TestColumnIndex(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"A", 0, false},
		{"B", 1, false},
		{"Z", 25, false},
		{"AA", 26, false},
		{"", 0, true},
		{"1", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := columnIndex(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestColumnName(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{-1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, columnName(tt.input))
		})
	}
}

func TestScanAll(t *testing.T) {
	filePath := createTestExcelFile(t)
	result, err := ScanAll[Person](filePath, "Test")
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func BenchmarkParse(b *testing.B) {
	row := []string{"Alice", "25"}
	for i := 0; i < b.N; i++ {
		_, _ = Parse[Person](row)
	}
}