package excel

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

const tagKey = "excel"

var structCache sync.Map // map[reflect.Type]structInfo

// IsXLSX reports whether the filename has .xlsx extension.
func IsXLSX(filename string) bool {
	return strings.EqualFold(filepath.Ext(filename), ".xlsx")
}

// Read reads all sheets from the Excel file.
// The returned map keys are sheet names.
//
// Example:
//
//	data, err := Read("test.xlsx")
//	for name, rows := range data {
//	    for _, row := range rows {
//	        fmt.Println(name, row)
//	    }
//	}
func Read(path string) (map[string][][]string, error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer wb.Close()
	return wb.ReadAll()
}

// ReadSheet reads a single sheet from the Excel file.
//
// Example:
//
//	rows, err := ReadSheet("test.xlsx", "Sheet1")
//	for _, row := range rows {
//	    fmt.Println(row)
//	}
func ReadSheet(path, name string) ([][]string, error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer wb.Close()
	return wb.Sheet(name).Rows()
}

// Walk iterates over rows and calls fn.
// If fn returns an error, iteration stops.
// fn receives the row index (1-based) and the row values.
//
// Example:
//
//	err := Walk("test.xlsx", "Sheet1", func(idx int, row []string) error {
//	    fmt.Println(idx, row)
//	    return nil
//	})
func Walk(path, name string, fn func(int, []string) error) error {
	wb, err := Open(path)
	if err != nil {
		return err
	}
	defer wb.Close()
	return wb.Sheet(name).Scan(fn)
}

// ScanRow reads the sheet and calls fn for each row.
// Parsing errors are skipped. If fn returns an error, the scan stops.
// fn receives the row index (1-based) and the parsed value.
//
// Example:
//
//	type Person struct {
//	    Name string `excel:"A"`
//	    Age  int    `excel:"B"`
//	}
//
//	err := ScanRow[Person]("test.xlsx", "Sheet1", func(idx int, p *Person) error {
//	    fmt.Println(idx, p.Name)
//	    return nil
//	})
func ScanRow[T any](path, name string, fn func(int, *T) error) error {
	wb, err := Open(path)
	if err != nil {
		return err
	}
	defer wb.Close()

	return wb.Sheet(name).Scan(func(idx int, row []string) error {
		v, err := Parse[T](row)
		if err != nil {
			return nil
		}
		return fn(idx, v)
	})
}

// ScanAll reads the sheet and returns all rows parsed as T.
// Rows that fail to parse are returned as nil.
//
// Example:
//
//	people, err := ScanAll[Person]("test.xlsx", "Sheet1")
func ScanAll[T any](path, name string) ([]*T, error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer wb.Close()

	var result []*T
	wb.Sheet(name).Scan(func(_ int, row []string) error {
		v, _ := Parse[T](row)
		result = append(result, v)
		return nil
	})
	return result, nil
}

// Parse parses a row into type T.
// Rows with fewer columns than required fields are handled gracefully.
//
// Example:
//
//	person, err := Parse[Person]([]string{"Alice", "25"})
func Parse[T any](row []string) (*T, error) {
	info := getStructInfo[T]()
	v := reflect.New(info.typ).Elem()

	for colIdx, f := range info.fields {
		if colIdx >= len(row) {
			continue
		}
		if err := setField(v.Field(f.index), row[colIdx]); err != nil {
			return nil, fmt.Errorf("column %s: %w", columnName(colIdx), err)
		}
	}

	result := new(T)
	reflect.ValueOf(result).Elem().Set(v)
	return result, nil
}

// Workbook represents an Excel workbook.
type Workbook struct {
	path string
	file *excelize.File
}

// Open opens an Excel file.
//
// Example:
//
//	wb, err := Open("test.xlsx")
//	if err != nil {
//	    return err
//	}
//	defer wb.Close()
func Open(path string) (*Workbook, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &Workbook{path: path, file: f}, nil
}

// Close closes the workbook.
func (w *Workbook) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Sheets returns all sheet names.
//
// Example:
//
//	names := wb.Sheets()
func (w *Workbook) Sheets() []string {
	if w.file == nil {
		return []string{}
	}
	return w.file.GetSheetList()
}

// Sheet returns a Sheet by name.
func (w *Workbook) Sheet(name string) *Sheet {
	return &Sheet{file: w.file, name: name}
}

// ReadAll reads all sheets and returns a map of sheet names to rows.
//
// Example:
//
//	data, err := wb.ReadAll()
func (w *Workbook) ReadAll() (map[string][][]string, error) {
	sheets := w.Sheets()
	result := make(map[string][][]string, len(sheets))

	for _, sheet := range sheets {
		rows, err := w.Sheet(sheet).Rows()
		if err != nil {
			return nil, err
		}
		result[sheet] = rows
	}

	return result, nil
}

// Sheet represents an Excel sheet.
type Sheet struct {
	file *excelize.File
	name string
}

// Rows returns all rows as [][]string.
func (s *Sheet) Rows() ([][]string, error) {
	return s.file.GetRows(s.name)
}

// Scan iterates over rows and calls fn.
// If fn returns an error, iteration stops.
func (s *Sheet) Scan(fn func(idx int, row []string) error) error {
	rows, err := s.file.Rows(s.name)
	if err != nil {
		return err
	}
	defer rows.Close()

	i := 1
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return err
		}
		if err := fn(i, cols); err != nil {
			return err
		}
		i++
	}
	return rows.Error()
}

// Row represents a row of cell values.
type Row struct {
	values []string
	index  int
}

// Values returns the cell values.
func (r *Row) Values() []string {
	return r.values
}

// Index returns the row number (1-based).
func (r *Row) Index() int {
	return r.index
}

// Value returns the value at column col (0-based).
// Returns empty string if col is out of range.
func (r *Row) Value(col int) string {
	if col < 0 || col >= len(r.values) {
		return ""
	}
	return r.values[col]
}

// Len returns the number of cells.
func (r *Row) Len() int {
	return len(r.values)
}

func getStructInfo[T any]() structInfo {
	var t T

	typ := reflect.TypeOf(t)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if info, ok := structCache.Load(typ); ok {
		return info.(structInfo)
	}

	info := structInfo{
		typ:    typ,
		fields: buildFieldIndex(typ),
	}
	structCache.Store(typ, info)
	return info
}

func buildFieldIndex(typ reflect.Type) map[int]fieldInfo {
	index := make(map[int]fieldInfo)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tag := field.Tag.Get(tagKey)
		if tag == "" || tag == "-" {
			continue
		}

		col, err := columnIndex(tag)
		if err != nil {
			continue
		}

		index[col] = fieldInfo{index: i}
	}

	return index
}

func columnIndex(col string) (int, error) {
	col = strings.ToUpper(col)

	if col == "" {
		return 0, errors.New("column name is empty")
	}

	n := 0
	for _, c := range col {
		if c < 'A' || c > 'Z' {
			return 0, fmt.Errorf("invalid column: %q", col)
		}
		n = n*26 + int(c-'A'+1)
	}

	return n - 1, nil
}

func columnName(n int) string {
	if n < 0 {
		return ""
	}

	var s string
	for n >= 0 {
		s = string(rune('A'+n%26)) + s
		n = n/26 - 1
		if n < 0 {
			break
		}
	}

	return s
}

func setField(v reflect.Value, s string) error {
	if !v.CanSet() {
		return nil
	}

	if v.Kind() == reflect.Ptr {
		if s == "" {
			v.Set(reflect.Zero(v.Type()))
			return nil
		}

		ptr := reflect.New(v.Type().Elem())

		if err := parseValue(ptr.Elem(), s); err != nil {
			return err
		}

		v.Set(ptr)
		return nil
	}

	return parseValue(v, s)
}

func parseValue(v reflect.Value, s string) error {
	if s == "" {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("parse int %q: %w", s, err)
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", s, err)
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("parse float %q: %w", s, err)
		}
		v.SetFloat(n)

	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", s, err)
		}
		v.SetBool(n)

	default:
		return fmt.Errorf("unsupported type: %v", v.Type())
	}

	return nil
}

type fieldInfo struct {
	index int
}

type structInfo struct {
	typ    reflect.Type
	fields map[int]fieldInfo
}
