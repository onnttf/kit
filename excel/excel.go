package excel

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

const tagKey = "excel"

var structCache sync.Map

// IsXLSX reports whether filename has an .xlsx extension.
func IsXLSX(filename string) bool {
	return strings.EqualFold(filepath.Ext(filename), ".xlsx")
}

// Read opens path and returns all sheets keyed by sheet name.
func Read(path string) (data map[string][][]string, err error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close workbook: %w", closeErr)
		}
	}()
	return wb.ReadAll()
}

// ReadSheet opens path and returns rows from one sheet.
func ReadSheet(path, name string) (rows [][]string, err error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close workbook: %w", closeErr)
		}
	}()
	return wb.Sheet(name).Rows()
}

// Walk opens path and calls fn for each row in one sheet.
func Walk(path, name string, fn func(int, []string) error) (err error) {
	wb, err := Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close workbook: %w", closeErr)
		}
	}()
	return wb.Sheet(name).Scan(fn)
}

// ScanRow parses each row into T and calls fn for rows that parse successfully.
func ScanRow[T any](path, name string, fn func(int, *T) error) (err error) {
	wb, err := Open(path)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close workbook: %w", closeErr)
		}
	}()

	return wb.Sheet(name).Scan(func(idx int, row []string) error {
		v, err := Parse[T](row)
		if err != nil {
			return nil
		}
		return fn(idx, v)
	})
}

// ScanAll parses all rows in a sheet into T. Rows that fail to parse are nil.
func ScanAll[T any](path, name string) (result []*T, err error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close workbook: %w", closeErr)
		}
	}()

	err = wb.Sheet(name).Scan(func(_ int, row []string) error {
		v, _ := Parse[T](row)
		result = append(result, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Parse maps row cells into T using struct fields tagged with excel column names.
func Parse[T any](row []string) (*T, error) {
	info, err := getStructInfo[T]()
	if err != nil {
		return nil, err
	}
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
	dst := reflect.ValueOf(result).Elem()
	if info.ptr {
		dst.Set(v.Addr())
		return result, nil
	}
	dst.Set(v)
	return result, nil
}

// Workbook wraps an opened Excel file.
type Workbook struct {
	path string
	file *excelize.File
}

// Open opens an Excel workbook at path.
func Open(path string) (*Workbook, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &Workbook{path: path, file: f}, nil
}

// Close closes the underlying workbook file.
func (w *Workbook) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Sheets returns all sheet names in the workbook.
func (w *Workbook) Sheets() []string {
	if w.file == nil {
		return []string{}
	}
	return w.file.GetSheetList()
}

// Sheet returns a sheet handle by name.
func (w *Workbook) Sheet(name string) *Sheet {
	return &Sheet{file: w.file, name: name}
}

// ReadAll returns all workbook sheets keyed by sheet name.
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

// Sheet provides row access for one workbook sheet.
type Sheet struct {
	file *excelize.File
	name string
}

// Rows returns all rows in the sheet.
func (s *Sheet) Rows() ([][]string, error) {
	return s.file.GetRows(s.name)
}

// Scan streams sheet rows and stops when fn returns an error.
func (s *Sheet) Scan(fn func(idx int, row []string) error) (err error) {
	rows, err := s.file.Rows(s.name)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close rows: %w", closeErr)
		}
	}()

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

// Row represents one row of cell values.
type Row struct {
	values []string
	index  int
}

// Values returns a defensive copy of the row values.
func (r *Row) Values() []string {
	return slices.Clone(r.values)
}

// Index returns the 1-based row index.
func (r *Row) Index() int {
	return r.index
}

// Value returns the cell at zero-based column col, or an empty string if missing.
func (r *Row) Value(col int) string {
	if col < 0 || col >= len(r.values) {
		return ""
	}
	return r.values[col]
}

// Len returns the number of values in the row.
func (r *Row) Len() int {
	return len(r.values)
}

type fieldInfo struct {
	index int
}

type structInfo struct {
	typ    reflect.Type
	ptr    bool
	fields map[int]fieldInfo
}

func getStructInfo[T any]() (structInfo, error) {
	var t T

	typ := reflect.TypeOf(t)
	if typ == nil {
		return structInfo{}, errors.New("excel: target must be struct or pointer to struct")
	}

	ptr := false
	elem := typ
	if typ.Kind() == reflect.Ptr {
		ptr = true
		elem = typ.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return structInfo{}, fmt.Errorf("excel: target must be struct or pointer to struct, got %v", typ)
	}

	if info, ok := structCache.Load(typ); ok {
		return info.(structInfo), nil
	}

	fields, err := buildFieldIndex(elem)
	if err != nil {
		return structInfo{}, err
	}

	info := structInfo{
		typ:    elem,
		ptr:    ptr,
		fields: fields,
	}
	structCache.Store(typ, info)
	return info, nil
}

func buildFieldIndex(typ reflect.Type) (map[int]fieldInfo, error) {
	index := make(map[int]fieldInfo)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tag := field.Tag.Get(tagKey)
		if tag == "" || tag == "-" {
			continue
		}

		col, err := columnIndex(tag)
		if err != nil {
			return nil, fmt.Errorf("field %s: %w", field.Name, err)
		}
		if field.PkgPath != "" {
			return nil, fmt.Errorf("field %s: field is unexported", field.Name)
		}

		index[col] = fieldInfo{index: i}
	}

	return index, nil
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
		if v.OverflowInt(n) {
			return fmt.Errorf("parse int %q: overflows %v", s, v.Type())
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", s, err)
		}
		if v.OverflowUint(n) {
			return fmt.Errorf("parse uint %q: overflows %v", s, v.Type())
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return fmt.Errorf("parse float %q: %w", s, err)
		}
		if v.OverflowFloat(n) {
			return fmt.Errorf("parse float %q: overflows %v", s, v.Type())
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
