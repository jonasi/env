package decode

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultSeparator      = "::"
	DefaultSliceSeparator = ","
)

var (
	DefaultKeyFunc = KeyIdentity
)

var (
	ErrInvalidKind = errors.New("Dest value must be a struct pointer")
)

type Options struct {
	Prefix         string
	Separator      string
	SliceSeparator string
	KeyFunc        func(string) string
}

func setDefaults(o *Options) {
	o.Separator = DefaultSeparator
	o.SliceSeparator = DefaultSliceSeparator
	o.KeyFunc = DefaultKeyFunc
}

func DecodeEnv(dest interface{}, opts *Options) error {
	return Decode(os.Environ(), dest, opts)
}

func Decode(args []string, dest interface{}, opts *Options) error {
	rootVal := reflect.ValueOf(dest)

	if rootVal.Kind() != reflect.Ptr {
		return ErrInvalidKind
	}

	rootVal = rootVal.Elem()

	if rootVal.Kind() != reflect.Struct {
		return ErrInvalidKind
	}

	if opts == nil {
		opts = &Options{}
	}

	setDefaults(opts)

	argsMap := argsMap(args)
	decodeStruct(rootVal, argsMap, opts, opts.Prefix)

	return nil
}

func decodeStruct(val reflect.Value, argsMap map[string]string, opts *Options, prefix string) {
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := valType.Field(i)

		// unexported
		if field.PkgPath != "" {
			continue
		}

		fieldVal := val.Field(i)

		k := opts.KeyFunc(field.Name)

		if v, ok := argsMap[k]; ok {
			decodeValue(fieldVal, v, opts)
		}
	}
}

func decodeValue(val reflect.Value, str string, opts *Options) {
	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type()))
		}

		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Slice:
		parts := strings.Split(str, opts.SliceSeparator)
		sliceVal := reflect.MakeSlice(val.Elem().Type(), len(parts), len(parts))

		for i, p := range parts {
			p = strings.TrimSpace(p)
			decodeValue(sliceVal.Index(i), p, opts)
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
}

func argsMap(args []string) map[string]string {
	m := map[string]string{}

	for _, a := range args {
		parts := strings.SplitN(a, "=", 2)

		if len(parts) == 1 {
			m[parts[0]] = ""
		} else {
			m[parts[0]] = parts[1]
		}
	}

	return m
}

func KeyIdentity(str string) string {
	return str
}

func ToUnderscore(str string) string {
	cp := str

	for _, c := range str {
		if unicode.IsUpper(c) {

		}
	}

	return cp
}
