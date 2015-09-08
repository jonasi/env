package env

import (
	"bufio"
	"bytes"
	"encoding"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// UnmarshalEnv parses os.Environ and stores the results in
// the value pointed to by dest
func UnmarshalEnv(dest interface{}, opts *Options) error {
	r := NewSliceReader(os.Environ())
	return unmarshal(r, dest, opts)
}

// Unmarshal parses data and stores the results in the value
// pointed to by dest
func Unmarshal(data []byte, dest interface{}, opts *Options) error {
	r := bytes.NewReader(data)
	return unmarshal(r, dest, opts)
}

func unmarshal(r io.Reader, dest interface{}, opts *Options) error {
	d := NewDecoder(r, opts)

	for d.More() {
		if err := d.Decode(dest); err != nil {
			return err
		}
	}

	return nil
}

// NewDecoder returns a new decoder that reads from r
func NewDecoder(r io.Reader, opts *Options) *Decoder {
	if opts == nil {
		opts = &Options{}
	}

	opts.setDefaults()

	return &Decoder{
		r:    bufio.NewReader(r),
		opts: opts,
		more: true,
	}
}

// A Decoder read k=v pairs from an input stream
type Decoder struct {
	r    *bufio.Reader
	opts *Options
	more bool
}

// More returns true if we have reached EOF on
// the underlying reader
func (d *Decoder) More() bool {
	return d.more
}

// Decode reads the next k=v pair from the input
// stream and stores the results in dest
func (d *Decoder) Decode(dest interface{}) error {
	rootVal := reflect.ValueOf(dest)

	if rootVal.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected pointer to struct, got %s", rootVal.Kind())
	}

	rootVal = rootVal.Elem()

	if rootVal.Kind() != reflect.Struct {
		return fmt.Errorf("Expected pointer to struct, got pointer to %s", rootVal.Kind())
	}

	str, err := d.r.ReadString('\n')

	if len(str) > 0 {
		if str[len(str)-1] == '\n' {
			drop := 1
			if len(str) > 1 && str[len(str)-2] == '\r' {
				drop = 2
			}

			str = str[:len(str)-drop]
		}
	}

	if err != nil {
		if err == io.EOF {
			d.more = false
			return d.decodeValue(rootVal, str)
		}

		return err
	}

	if err := d.decodeValue(rootVal, str); err != nil {
		return err
	}

	return nil
}

func (d *Decoder) decodeValue(rv reflect.Value, str string) error {
	var (
		parts = strings.SplitN(str, "=", 2)
		key   = parts[0]
		v     string
	)

	if d.opts.Prefix != "" {
		if !strings.HasPrefix(key, d.opts.Prefix) {
			return nil
		}

		key = key[len(d.opts.Prefix):]
	}

	if len(parts) > 1 {
		v = parts[1]
	}

	v = strings.TrimSpace(v)

	var (
		kparts  = strings.Split(key, d.opts.Separator)
		cur, ok = deref(rv)
	)

	if !ok {
		return nil
	}

	for _, k := range kparts {
		f, ok := cur.Type().FieldByNameFunc(func(f string) bool {
			return d.opts.Mapper(f) == k
		})

		if !ok {
			return nil
		}

		if f.PkgPath != "" {
			return nil
		}

		cur, ok = deref(cur.FieldByIndex(f.Index))

		if !ok {
			return nil
		}
	}

	return decodeString(v, cur, d.opts)
}

// DecodeString parses string for a go value and stores it in dest
func DecodeString(str string, dest interface{}, opts *Options) error {
	val := reflect.ValueOf(dest)
	return decodeString(str, val, opts)
}

func decodeString(str string, val reflect.Value, opts *Options) error {
	if checkUnmarshaler(val, str) {
		return nil
	}

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}

		if checkUnmarshaler(val, str) {
			return nil
		}

		val = val.Elem()
	}

	if checkUnmarshaler(val, str) {
		return nil
	}

	switch val.Kind() {
	case reflect.Slice:
		str = strings.TrimSpace(str)
		if str == "" {
			val.Set(reflect.MakeSlice(val.Type(), 0, 0))
			return nil
		}

		parts := strings.Split(str, opts.SliceSeparator)
		sliceVal := reflect.MakeSlice(val.Type(), len(parts), len(parts))

		for i, p := range parts {
			p = strings.TrimSpace(p)
			decodeString(p, sliceVal.Index(i), opts)
		}

		val.Set(sliceVal)
	case reflect.String:
		val.SetString(str)
	case reflect.Bool:
		v, _ := strconv.ParseBool(str)
		val.SetBool(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, _ := strconv.ParseInt(str, 0, val.Type().Bits())
		val.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, _ := strconv.ParseUint(str, 0, val.Type().Bits())
		val.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, _ := strconv.ParseFloat(str, val.Type().Bits())
		val.SetFloat(v)
	}

	return nil
}

func deref(v reflect.Value) (reflect.Value, bool) {
	for {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}

			v = v.Elem()
		} else if v.Kind() == reflect.Interface {
			if v.IsNil() {
				return v, false
			}

			v = v.Elem()
		} else {
			break
		}
	}

	return v, true
}

var unmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func checkUnmarshaler(val reflect.Value, str string) bool {
	if val.Type().Implements(unmarshaler) {
		val.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
		return true
	}

	return false
}
