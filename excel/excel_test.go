package excel

import (
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/xuri/excelize/v2"
)

type Person struct {
	Name string `excel:"A"`
	Age  int    `excel:"B"`
}

func TestOpenRowsScanDecode(t *testing.T) {
	path := createWorkbook(t)
	wb, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer closeWorkbook(t, wb)

	if got := wb.Sheets(); !reflect.DeepEqual(got, []string{"Sheet1"}) {
		t.Fatalf("Sheets() = %v", got)
	}

	rows, err := wb.Rows("Sheet1")
	if err != nil {
		t.Fatalf("Rows() error = %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("Rows() len = %d", len(rows))
	}

	var seen []int
	if err := wb.Sheet("Sheet1").Scan(func(index int, row []string) error {
		seen = append(seen, index)
		return nil
	}); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if !reflect.DeepEqual(seen, []int{1, 2}) {
		t.Fatalf("Scan indexes = %v", seen)
	}

	people, err := Decode[Person](wb.Sheet("Sheet1"))
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(people) != 2 || people[0].Name != "alice" || people[1].Age != 40 {
		t.Fatalf("Decode() = %+v", people)
	}
}

func TestReadAndDecodeError(t *testing.T) {
	path := createWorkbook(t)
	people, err := Read[Person](path, "Sheet1")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if len(people) != 2 {
		t.Fatalf("Read() len = %d", len(people))
	}

	f := excelize.NewFile()
	t.Cleanup(func() { _ = f.Close() })
	setCellValue(t, f, "A1", "bad")
	setCellValue(t, f, "B1", "not-int")
	bad := t.TempDir() + "/bad.xlsx"
	if err := f.SaveAs(bad); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	_, err = Read[Person](bad, "Sheet1")
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("Read(bad) error = %v", err)
	}
}

func TestOpenReaderAndInvalidInput(t *testing.T) {
	path := createWorkbook(t)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open file: %v", err)
	}
	defer closeFile(t, file)

	wb, err := OpenReader(file)
	if err != nil {
		t.Fatalf("OpenReader() error = %v", err)
	}
	defer closeWorkbook(t, wb)

	if err := wb.Sheet("Sheet1").Scan(nil); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Scan(nil) error = %v", err)
	}
}

func createWorkbook(t *testing.T) string {
	t.Helper()
	f := excelize.NewFile()
	t.Cleanup(func() { _ = f.Close() })
	setCellValue(t, f, "A1", "alice")
	setCellValue(t, f, "B1", 30)
	setCellValue(t, f, "A2", "bob")
	setCellValue(t, f, "B2", 40)
	path := t.TempDir() + "/people.xlsx"
	if err := f.SaveAs(path); err != nil {
		t.Fatalf("SaveAs() error = %v", err)
	}
	return path
}

func setCellValue(t testing.TB, f *excelize.File, cell string, value any) {
	t.Helper()
	if err := f.SetCellValue("Sheet1", cell, value); err != nil {
		t.Fatalf("SetCellValue(%q) error = %v", cell, err)
	}
}

func closeWorkbook(t testing.TB, wb *Workbook) {
	t.Helper()
	if err := wb.Close(); err != nil {
		t.Fatalf("Close workbook error = %v", err)
	}
}

func closeFile(t testing.TB, file *os.File) {
	t.Helper()
	if err := file.Close(); err != nil {
		t.Fatalf("Close file error = %v", err)
	}
}

type allTypesRow struct {
	S string  `excel:"A"`
	B bool    `excel:"B"`
	I int     `excel:"C"`
	U uint16  `excel:"D"`
	F float64 `excel:"E"`
}

func TestDecodePointerAndScalarTypes(t *testing.T) {
	got, err := decodeRow[*allTypesRow]([]string{
		"hello",
		"true",
		"-8",
		"16",
		"3.5",
	})
	if err != nil {
		t.Fatalf("decodeRow() error = %v", err)
	}
	want := &allTypesRow{
		S: "hello",
		B: true,
		I: -8,
		U: 16,
		F: 3.5,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("decodeRow() = %+v, want %+v", got, want)
	}
}

func TestDecodeSkipsMissingAndEmptyCells(t *testing.T) {
	got, err := decodeRow[allTypesRow]([]string{"only-name", ""})
	if err != nil {
		t.Fatalf("decodeRow() error = %v", err)
	}
	if got.S != "only-name" || got.B || got.I != 0 || got.U != 0 || got.F != 0 {
		t.Fatalf("decodeRow() = %+v", got)
	}
}

func TestDecodeInvalidTargetsAndTags(t *testing.T) {
	if _, err := decodeRow[int]([]string{"1"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("decodeRow[int]() error = %v", err)
	}

	type privateField struct {
		name string `excel:"A"`
	}
	private := privateField{name: "x"}
	_ = private.name
	if _, err := decodeRow[privateField]([]string{"x"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("decodeRow[privateField]() error = %v", err)
	}

	type invalidColumn struct {
		Name string `excel:"A1"`
	}
	if _, err := decodeRow[invalidColumn]([]string{"x"}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("decodeRow[invalidColumn]() error = %v", err)
	}

	type unsupported struct {
		Value complex64 `excel:"A"`
	}
	if _, err := decodeRow[unsupported]([]string{"1"}); !errors.Is(err, ErrDecode) {
		t.Fatalf("decodeRow[unsupported]() error = %v", err)
	}
}

func TestColumnIndexAndName(t *testing.T) {
	tests := []struct {
		col  string
		idx  int
		name string
	}{
		{col: "A", idx: 0, name: "A"},
		{col: "Z", idx: 25, name: "Z"},
		{col: "AA", idx: 26, name: "AA"},
		{col: " ab ", idx: 27, name: "AB"},
	}

	for _, tt := range tests {
		t.Run(tt.col, func(t *testing.T) {
			got, err := columnIndex(tt.col)
			if err != nil {
				t.Fatalf("columnIndex() error = %v", err)
			}
			if got != tt.idx {
				t.Fatalf("columnIndex() = %d, want %d", got, tt.idx)
			}
			if name := columnName(got); name != tt.name {
				t.Fatalf("columnName() = %q, want %q", name, tt.name)
			}
		})
	}

	if _, err := columnIndex(" "); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("columnIndex(empty) error = %v", err)
	}
}

func TestNilWorkbookAndSheet(t *testing.T) {
	var wb *Workbook
	if err := wb.Close(); err != nil {
		t.Fatalf("nil Close() error = %v", err)
	}
	if got := wb.Sheets(); got != nil {
		t.Fatalf("nil Sheets() = %v", got)
	}
	if _, err := wb.Rows("Sheet1"); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil Rows() error = %v", err)
	}

	var sheet *Sheet
	if _, err := sheet.Rows(); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil sheet Rows() error = %v", err)
	}
	if err := sheet.Scan(func(int, []string) error { return nil }); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("nil sheet Scan() error = %v", err)
	}
}
