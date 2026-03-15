package excel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestIsXLSX(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"xlsx extension", "data.xlsx", true},
		{"XLSX upper", "data.XLSX", true},
		{"xls extension", "data.xls", false},
		{"csv extension", "data.csv", false},
		{"no extension", "data", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsXLSX(tt.filename))
		})
	}
}

func TestRead(t *testing.T) {
	testFile := createTestFile(t)
	defer os.Remove(testFile)

	allSheets, err := Read(testFile)
	require.NoError(t, err)
	assert.Contains(t, allSheets, "Sheet1")
	assert.Equal(t, 3, len(allSheets["Sheet1"]))
	assert.Equal(t, []string{"Name", "Age", "City"}, allSheets["Sheet1"][0])
}

func TestReadSheet(t *testing.T) {
	testFile := createTestFileWithData(t, [][]string{
		{"Name", "Age"},
		{"Alice", "25"},
		{"Bob", "30"},
	})
	defer os.Remove(testFile)

	rows, err := ReadSheet(testFile, "Sheet1")
	require.NoError(t, err)
	assert.Equal(t, 3, len(rows))
	assert.Equal(t, []string{"Name", "Age"}, rows[0])
	assert.Equal(t, []string{"Alice", "25"}, rows[1])
}

func TestScanSheet(t *testing.T) {
	type Person struct {
		Name string `excel:"A"`
		Age  int    `excel:"B"`
		City string `excel:"C"`
	}

	testFile := createTestFileWithData(t, [][]string{
		{"Name", "Age", "City"},
		{"Alice", "25", "Beijing"},
		{"Bob", "30", "Shanghai"},
	})
	defer os.Remove(testFile)

	people, err := ScanSheet[Person](testFile, "Sheet1")
	require.NoError(t, err)

	// First row (header) fails to parse, so it becomes nil
	assert.Equal(t, 3, len(people))
	// Filter out nil values
	var dataRows []*Person
	for _, p := range people {
		if p != nil {
			dataRows = append(dataRows, p)
		}
	}

	assert.Equal(t, 2, len(dataRows))
	assert.Equal(t, "Alice", dataRows[0].Name)
	assert.Equal(t, 25, dataRows[0].Age)
	assert.Equal(t, "Bob", dataRows[1].Name)
	assert.Equal(t, 30, dataRows[1].Age)
}

func TestScanSheetEmptyFile(t *testing.T) {
	type Person struct {
		Name string `excel:"A"`
	}

	testFile := createTestFileWithData(t, [][]string{
		{"Name"},
	})
	defer os.Remove(testFile)

	people, err := ScanSheet[Person](testFile, "Sheet1")
	require.NoError(t, err)
	// Header only, fails to parse
	assert.Equal(t, 1, len(people))
}

func TestWalk(t *testing.T) {
	testFile := createTestFileWithData(t, [][]string{
		{"Name", "Age"},
		{"Alice", "25"},
		{"Bob", "30"},
	})
	defer os.Remove(testFile)

	var rows [][]string
	err := Walk(testFile, "Sheet1", func(row []string) error {
		rows = append(rows, row)
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, 3, len(rows))
	assert.Equal(t, []string{"Name", "Age"}, rows[0])
	assert.Equal(t, []string{"Alice", "25"}, rows[1])
	assert.Equal(t, []string{"Bob", "30"}, rows[2])
}

func TestScan(t *testing.T) {
	type Person struct {
		Name string `excel:"A"`
		Age  int    `excel:"B"`
	}

	testFile := createTestFileWithData(t, [][]string{
		{"Name", "Age"},
		{"Alice", "25"},
		{"Bob", "30"},
	})
	defer os.Remove(testFile)

	var people []*Person
	err := Scan(testFile, "Sheet1", func(p *Person) error {
		people = append(people, p)
		return nil
	})
	require.NoError(t, err)

	// Header row is skipped because it fails to parse
	assert.Equal(t, 2, len(people))
	assert.Equal(t, "Alice", people[0].Name)
	assert.Equal(t, 25, people[0].Age)
	assert.Equal(t, "Bob", people[1].Name)
	assert.Equal(t, 30, people[1].Age)
}

// Helper function to create a test Excel file
func createTestFile(t *testing.T) string {
	return createTestFileWithData(t, [][]string{
		{"Name", "Age", "City"},
		{"Alice", "25", "Beijing"},
		{"Bob", "30", "Shanghai"},
	})
}

// Helper function to create a test Excel file with custom data
func createTestFileWithData(t *testing.T, data [][]string) string {
	f := excelize.NewFile()
	defer f.Close()

	for rowIdx, row := range data {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue("Sheet1", cell, value)
		}
	}

	tmpFile := t.TempDir() + "/test.xlsx"
	err := f.SaveAs(tmpFile)
	require.NoError(t, err)

	return tmpFile
}
