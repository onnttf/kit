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

const excelTag = "excel"

var structCache sync.Map // map[reflect.Type]structInfo

// IsXLSX reports whether the file has an .xlsx extension.
func IsXLSX(filename string) bool {
	return strings.EqualFold(filepath.Ext(filename), ".xlsx")
}

type structField struct {
	index int
}

type structInfo struct {
	typ    reflect.Type
	fields map[int]structField
}

// Read reads all sheets from the Excel file at path.
// The returned map keys are sheet names.
//
// Example:
//
//	data, err := excel.Read("test.xlsx")
//	for sheetName, rows := range data {
//	    fmt.Println("Sheet:", sheetName)
//	    for _, row := range rows {
//	        fmt.Println(row)
//	    }
//	}
func Read(path string) (map[string][][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sheets := f.GetSheetList()
	result := make(map[string][][]string, len(sheets))

	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, err
		}
		result[sheet] = rows
	}

	return result, nil
}

// ReadSheet reads a single sheet from the Excel file.
//
// Example:
//
//	rows, err := excel.ReadSheet("test.xlsx", "Sheet1")
//	for _, row := range rows {
//	    fmt.Println(row)
//	}
func ReadSheet(path, sheet string) ([][]string, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.GetRows(sheet)
}

// Walk reads the Excel file row by row and calls fn for each row.
// It avoids loading the entire sheet into memory.
//
// Example:
//
//	err := excel.Walk("test.xlsx", "Sheet1", func(index int, row []string) error {
//	    fmt.Println(index, row)
//	    return nil
//	})
func Walk(path, sheet string, fn func(index int, row []string) error) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	index := 1
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		if err := fn(index, columns); err != nil {
			return err
		}
		index++
	}

	return rows.Error()
}

// Scan reads the Excel file row by row, unmarshals each row into T,
// and calls fn for each row.
//
// Rows that fail to parse are skipped.
//
// Example:
//
//	type Person struct {
//	    Name string `excel:"A"`
//	    Age  int    `excel:"B"`
//	}
//
//	err := excel.Scan("test.xlsx", "Sheet1", func(index int, row *Person) error {
//	    fmt.Printf("Row %d: Name: %s, Age: %d\n", index, row.Name, row.Age)
//	    return nil
//	})
func Scan[T any](path, sheet string, fn func(index int, row *T) error) error {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	info := getStructInfo[T]()

	index := 1
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return err
		}

		v, err := unmarshalRow(columns, info.typ, info.fields, index)
		if err != nil {
			index++
			continue
		}

		t := v.Interface().(T)

		if err := fn(index, &t); err != nil {
			return err
		}
		index++
	}

	return rows.Error()
}

// ScanSheet reads all rows from sheet and returns them as a slice of T.
//
// Example:
//
//	type Person struct {
//	    Name string `excel:"A"`
//	    Age  int    `excel:"B"`
//	}
//
// Rows that fail to parse are returned as nil.
func ScanSheet[T any](path, sheet string) ([]*T, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, err
	}

	info := getStructInfo[T]()

	result := make([]*T, 0, len(rows))

	for i, row := range rows {
		v, err := unmarshalRow(row, info.typ, info.fields, i+1)
		if err != nil {
			result = append(result, nil)
			continue
		}

		t := v.Interface().(T)
		result = append(result, &t)
	}

	return result, nil
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

func buildFieldIndex(typ reflect.Type) map[int]structField {
	index := make(map[int]structField)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		tag := field.Tag.Get(excelTag)
		if tag == "" || tag == "-" {
			continue
		}

		col, err := columnIndex(tag)
		if err != nil {
			continue
		}

		index[col] = structField{index: i}
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

func unmarshalRow(
	columns []string,
	typ reflect.Type,
	fields map[int]structField,
	row int,
) (reflect.Value, error) {
	v := reflect.New(typ).Elem()

	for colIndex, f := range fields {
		if colIndex >= len(columns) {
			continue
		}

		if err := setField(v.Field(f.index), columns[colIndex]); err != nil {
			return reflect.Value{}, fmt.Errorf("row %d, column %s: %w", row, columnName(colIndex), err)
		}
	}

	return v, nil
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
