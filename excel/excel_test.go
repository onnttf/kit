package excel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

var testData = [][]string{
	{"Name", "Age", "City", "Score", "Active"},
	{"Alice", "25", "Beijing", "95.5", "true"},
	{"Bob", "30", "Shanghai", "88.0", "false"},
}

type Person struct {
	Name   string  `excel:"A"`
	Age    int     `excel:"B"`
	City   string  `excel:"C"`
	Score  float64 `excel:"D"`
	Active bool    `excel:"E"`
}

func createTestFile(t *testing.T, data [][]string) string {
	f := excelize.NewFile()
	defer f.Close()

	for rowIdx, row := range data {
		for colIdx, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+1)
			f.SetCellValue("Sheet1", cell, value)
		}
	}

	tmpFile := t.TempDir() + "/test.xlsx"
	require.NoError(t, f.SaveAs(tmpFile))
	return tmpFile
}

func TestIsXLSX(t *testing.T) {
	assert.True(t, IsXLSX("data.xlsx"))
	assert.True(t, IsXLSX("data.XLSX"))
	assert.False(t, IsXLSX("data.xls"))
	assert.False(t, IsXLSX("data.csv"))
}

func TestOpen(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	wb, err := Open(testFile)
	require.NoError(t, err)
	require.NoError(t, wb.Close())
}

func TestOpen_NotFound(t *testing.T) {
	_, err := Open("nonexistent.xlsx")
	assert.Error(t, err)
}

func TestWorkbook_Sheets(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	wb, _ := Open(testFile)
	defer wb.Close()

	sheets := wb.Sheets()
	assert.Contains(t, sheets, "Sheet1")
}

func TestRead(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	data, err := Read(testFile)
	require.NoError(t, err)
	assert.Equal(t, testData, data["Sheet1"])
}

func TestReadSheet(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	rows, err := ReadSheet(testFile, "Sheet1")
	require.NoError(t, err)
	assert.Equal(t, testData, rows)
}

func TestWalk(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	var rows [][]string
	err := Walk(testFile, "Sheet1", func(idx int, row []string) error {
		rows = append(rows, row)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, testData, rows)
}

func TestWalk_Stop(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	err := Walk(testFile, "Sheet1", func(idx int, row []string) error {
		if idx == 2 {
			return assert.AnError
		}
		return nil
	})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestScanRow(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	var people []*Person
	err := ScanRow(testFile, "Sheet1", func(idx int, p *Person) error {
		people = append(people, p)
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, 2, len(people))
	assert.Equal(t, "Alice", people[0].Name)
}

func TestScanRow_Stop(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	err := ScanRow(testFile, "Sheet1", func(idx int, p *Person) error {
		if idx == 2 {
			return assert.AnError
		}
		return nil
	})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestScanAll(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	people, err := ScanAll[Person](testFile, "Sheet1")
	require.NoError(t, err)

	assert.Equal(t, 3, len(people))
	assert.Nil(t, people[0]) // header
	assert.Equal(t, "Alice", people[1].Name)
}

func TestParse(t *testing.T) {
	row := []string{"Alice", "25", "Beijing", "95.5", "true"}
	p, err := Parse[Person](row)
	require.NoError(t, err)

	assert.Equal(t, "Alice", p.Name)
	assert.Equal(t, 25, p.Age)
	assert.Equal(t, 95.5, p.Score)
	assert.True(t, p.Active)
}

func TestParse_InvalidInt(t *testing.T) {
	row := []string{"Name", "not_a_number"}
	_, err := Parse[Person](row)
	assert.Error(t, err)
}

func TestSheet_Scan(t *testing.T) {
	testFile := createTestFile(t, testData)
	defer os.Remove(testFile)

	wb, _ := Open(testFile)
	defer wb.Close()

	var rows [][]string
	err := wb.Sheet("Sheet1").Scan(func(idx int, row []string) error {
		rows = append(rows, row)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, testData, rows)
}
