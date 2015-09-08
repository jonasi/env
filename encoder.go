package env

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Marshal returns the env encoding of v
func Marshal(v interface{}, opts *Options) ([]byte, error) {
	var (
		b bytes.Buffer
		e = NewEncoder(&b, opts)
	)

	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// NewEncoder returns a new encoder that writes to w
func NewEncoder(w io.Writer, opts *Options) *Encoder {
	if opts == nil {
		opts = &Options{}
	}

	opts.setDefaults()

	return &Encoder{
		w:    w,
		opts: opts,
	}
}

// An Encoder writes env k=v pairs to an output stream
type Encoder struct {
	w    io.Writer
	opts *Options
}

// Encode writes the env encoding to the stream
func (e *Encoder) Encode(src interface{}) error {
	rootVal := reflect.ValueOf(src)

	for rootVal.Kind() == reflect.Ptr || rootVal.Kind() == reflect.Interface {
		rootVal = rootVal.Elem()
	}

	return encodeStruct(e.w, rootVal, e.opts, e.opts.Prefix, true)
}

func encodeStruct(w io.Writer, rv reflect.Value, opts *Options, prefix string, isFirst bool) error {
	if rv.Kind() != reflect.Struct {
		panic("Expected struct, but found " + rv.Kind().String())
	}

	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if field.PkgPath != "" {
			continue
		}

		fieldVal := rv.Field(i)

		for fieldVal.Kind() == reflect.Ptr || fieldVal.Kind() == reflect.Interface {
			if fieldVal.IsNil() {
				continue
			}

			fieldVal = fieldVal.Elem()
		}

		k := prefix + opts.Mapper(field.Name)

		if fieldVal.Kind() != reflect.Struct {
			v := k + "=" + encodeString(fieldVal, opts)

			if !isFirst {
				v = "\n" + v
			}

			isFirst = false

			if _, err := w.Write([]byte(v)); err != nil {
				return err
			}
		} else {
			if err := encodeStruct(w, fieldVal, opts, k+opts.Separator, isFirst); err != nil {
				return err
			}
		}

		k = opts.Mapper(k)
	}
	return nil
}

// EncodeString encodes and returns val as a string
func EncodeString(val interface{}, opts *Options) string {
	v := reflect.ValueOf(val)
	return encodeString(v, opts)
}

func encodeString(v reflect.Value, opts *Options) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits())
	case reflect.Slice:
		l := v.Len()
		vals := make([]string, l)

		for i := 0; i < l; i++ {
			sv := v.Index(i)
			vals[i] = encodeString(sv, opts)
		}

		return strings.Join(vals, opts.SliceSeparator)
	}

	panic("EncodeString unsupported val kind " + v.Kind().String())
}
