package excel

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/xuri/excelize/v2"
)

const tagKey = "excel"

var (
	structCache sync.Map

	ErrInvalidInput = errors.New("excel: invalid input")
	ErrDecode       = errors.New("excel: decode error")
)

type Workbook struct {
	file *excelize.File
}

type Sheet struct {
	file *excelize.File
	name string
}

func Open(path string) (*Workbook, error) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("excel: open file: %w", err)
	}
	return &Workbook{file: f}, nil
}

func OpenReader(r io.Reader) (*Workbook, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return nil, fmt.Errorf("excel: open reader: %w", err)
	}
	return &Workbook{file: f}, nil
}

func Read[T any](path, sheet string) (result []T, err error) {
	wb, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := wb.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	return Decode[T](wb.Sheet(sheet))
}

func Decode[T any](s *Sheet) ([]T, error) {
	var out []T
	if err := s.Scan(func(_ int, row []string) error {
		v, err := decodeRow[T](row)
		if err != nil {
			return err
		}
		out = append(out, v)
		return nil
	}); err != nil {
		return nil, err
	}
	return out, nil
}

func (w *Workbook) Close() error {
	if w == nil || w.file == nil {
		return nil
	}
	if err := w.file.Close(); err != nil {
		return fmt.Errorf("excel: close workbook: %w", err)
	}
	return nil
}

func (w *Workbook) Sheets() []string {
	if w == nil || w.file == nil {
		return nil
	}
	return w.file.GetSheetList()
}

func (w *Workbook) Sheet(name string) *Sheet {
	if w == nil {
		return &Sheet{name: name}
	}
	return &Sheet{file: w.file, name: name}
}

func (w *Workbook) Rows(sheet string) ([][]string, error) {
	return w.Sheet(sheet).Rows()
}

func (s *Sheet) Rows() ([][]string, error) {
	if s == nil || s.file == nil {
		return nil, fmt.Errorf("%w: nil sheet", ErrInvalidInput)
	}
	rows, err := s.file.GetRows(s.name)
	if err != nil {
		return nil, fmt.Errorf("excel: get rows: %w", err)
	}
	return rows, nil
}

func (s *Sheet) Scan(fn func(index int, row []string) error) (err error) {
	if fn == nil {
		return ErrInvalidInput
	}
	if s == nil || s.file == nil {
		return fmt.Errorf("%w: nil sheet", ErrInvalidInput)
	}
	rows, err := s.file.Rows(s.name)
	if err != nil {
		return fmt.Errorf("excel: scan rows: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("excel: close rows: %w", closeErr)
		}
	}()

	i := 1
	for rows.Next() {
		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("excel: read row: %w", err)
		}
		if err := fn(i, cols); err != nil {
			return err
		}
		i++
	}
	if err := rows.Error(); err != nil {
		return fmt.Errorf("excel: scan rows: %w", err)
	}
	return nil
}

type fieldInfo struct {
	index int
	typ   reflect.Type
}

type structInfo struct {
	typ    reflect.Type
	ptr    bool
	fields map[int]fieldInfo
}

func decodeRow[T any](row []string) (T, error) {
	info, err := getStructInfo[T]()
	if err != nil {
		var zero T
		return zero, err
	}
	v := reflect.New(info.typ).Elem()

	for colIdx, f := range info.fields {
		if colIdx >= len(row) {
			continue
		}
		if err := setField(v.Field(f.index), row[colIdx]); err != nil {
			var zero T
			return zero, fmt.Errorf("%w: column=%s: %w", ErrDecode, columnName(colIdx), err)
		}
	}

	var result T
	dst := reflect.ValueOf(&result).Elem()
	if info.ptr {
		dst.Set(v.Addr())
		return result, nil
	}
	dst.Set(v)
	return result, nil
}

func getStructInfo[T any]() (structInfo, error) {
	var zero T
	typ := reflect.TypeOf(zero)
	ptr := false
	if typ == nil {
		return structInfo{}, fmt.Errorf("%w: nil target", ErrInvalidInput)
	}
	if typ.Kind() == reflect.Pointer {
		ptr = true
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return structInfo{}, fmt.Errorf("%w: target must be struct", ErrInvalidInput)
	}

	if cached, ok := structCache.Load(typ); ok {
		info := cached.(structInfo)
		info.ptr = ptr
		return info, nil
	}

	fields, err := buildFieldIndex(typ)
	if err != nil {
		return structInfo{}, err
	}
	info := structInfo{typ: typ, fields: fields}
	structCache.Store(typ, info)
	info.ptr = ptr
	return info, nil
}

func buildFieldIndex(typ reflect.Type) (map[int]fieldInfo, error) {
	fields := make(map[int]fieldInfo)
	for i := range typ.NumField() {
		field := typ.Field(i)
		tag := field.Tag.Get(tagKey)
		if tag == "" || tag == "-" {
			continue
		}
		if field.PkgPath != "" {
			return nil, fmt.Errorf("%w: field=%s", ErrInvalidInput, field.Name)
		}
		col, err := columnIndex(tag)
		if err != nil {
			return nil, err
		}
		fields[col] = fieldInfo{index: i, typ: field.Type}
	}
	return fields, nil
}

func columnIndex(col string) (int, error) {
	col = strings.TrimSpace(strings.ToUpper(col))
	if col == "" {
		return 0, fmt.Errorf("%w: empty column", ErrInvalidInput)
	}
	n := 0
	for _, r := range col {
		if r < 'A' || r > 'Z' {
			return 0, fmt.Errorf("%w: column=%q", ErrInvalidInput, col)
		}
		n = n*26 + int(r-'A'+1)
	}
	return n - 1, nil
}

func columnName(n int) string {
	n++
	var b []byte
	for n > 0 {
		n--
		b = append([]byte{byte('A' + n%26)}, b...)
		n /= 26
	}
	return string(b)
}

func setField(v reflect.Value, s string) error {
	if !v.CanSet() {
		return fmt.Errorf("%w: unsettable field", ErrInvalidInput)
	}
	if s == "" {
		return nil
	}
	return parseValue(v, s)
}

func parseValue(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		u, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("%w: type=%s", ErrInvalidInput, v.Type())
	}
	return nil
}
