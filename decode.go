package envdecode

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

const (
	DefaultSeparator      = "__"
	DefaultSliceSeparator = ","
)

var (
	DefaultKeyFunc = KeyIdentity
)

type Options struct {
	Prefix         string
	Separator      string
	SliceSeparator string
	KeyFunc        func(string) string
}

func setDefaults(o *Options) {
	if o.Separator == "" {
		o.Separator = DefaultSeparator
	}

	if o.SliceSeparator == "" {
		o.SliceSeparator = DefaultSliceSeparator
	}

	if o.KeyFunc == nil {
		o.KeyFunc = DefaultKeyFunc
	}
}

func DecodeEnv(dest interface{}, opts *Options) error {
	return Decode(os.Environ(), dest, opts)
}

func Decode(args []string, dest interface{}, opts *Options) error {
	rootVal := reflect.ValueOf(dest)

	if rootVal.Kind() != reflect.Ptr {
		return fmt.Errorf("Expected pointer to struct, got %s", rootVal.Kind())
	}

	rootVal = rootVal.Elem()

	if rootVal.Kind() != reflect.Struct {
		return fmt.Errorf("Expected pointer to struct, got pointer to %s", rootVal.Kind())
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

		if fieldVal.Kind() == reflect.Struct {
			decodeStruct(fieldVal, argsMap, opts, prefix+field.Name+opts.Separator)
			continue
		}

		if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				fieldVal = fieldVal.Elem()
			}

			decodeStruct(fieldVal, argsMap, opts, prefix+field.Name+opts.Separator)

			continue
		}

		k := opts.KeyFunc(prefix + field.Name)

		if v, ok := argsMap[k]; ok {
			decodeValue(fieldVal, v, opts)
		}
	}
}

var unmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func checkInterface(val reflect.Value, str string) bool {
	if val.Type().Implements(unmarshaler) {
		val.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
		return true
	}

	return false
}

// decode a string into a value
func decodeValue(val reflect.Value, str string, opts *Options) {
	if checkInterface(val, str) {
		return
	}

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}

		if checkInterface(val, str) {
			return
		}

		val = val.Elem()
	}

	if checkInterface(val, str) {
		return
	}

	switch val.Kind() {
	case reflect.Slice:
		str = strings.TrimSpace(str)
		if str == "" {
			val.Set(reflect.MakeSlice(val.Type(), 0, 0))
			return
		}

		parts := strings.Split(str, opts.SliceSeparator)
		sliceVal := reflect.MakeSlice(val.Type(), len(parts), len(parts))

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
	var (
		parts = []string{}
		cur   = []rune{}
		last2 = [2]rune{}
	)

	for _, c := range str {
		if unicode.IsUpper(c) {
			if last2[1] != 0 && unicode.IsLower(last2[1]) {
				parts = append(parts, string(cur))
				cur = nil
			}

			cur = append(cur, unicode.ToLower(c))
		} else {
			if last2[0] != 0 && last2[1] != 0 && unicode.IsUpper(last2[0]) && unicode.IsUpper(last2[1]) {
				parts = append(parts, string(cur[:len(cur)-1]))
				cur = []rune{cur[len(cur)-1]}
			}

			cur = append(cur, c)
		}

		last2[0] = last2[1]
		last2[1] = c
	}

	if len(cur) > 0 {
		parts = append(parts, string(cur))
	}

	return strings.Join(parts, "_")
}
