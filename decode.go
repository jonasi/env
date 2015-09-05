package env

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Default option values
const (
	DefaultSeparator      = "__"
	DefaultSliceSeparator = ","
)

// Default option values
var (
	DefaultMapper = IdentityMapper
)

// Options is the set of decode options
type Options struct {
	Prefix         string              // String prefix that is stripped from each entry
	Separator      string              // String value for nested values
	SliceSeparator string              // Separator for slices encoded as strings
	Mapper         func(string) string // Mapping function for struct key names
}

func setDefaults(o *Options) {
	if o.Separator == "" {
		o.Separator = DefaultSeparator
	}

	if o.SliceSeparator == "" {
		o.SliceSeparator = DefaultSliceSeparator
	}

	if o.Mapper == nil {
		o.Mapper = DefaultMapper
	}
}

// DecodeEnv calls `Decode` with `os.Args`
func UnmarshalEnv(dest interface{}, opts *Options) error {
	return Unmarshal(os.Environ(), dest, opts)
}

// Decode decodes the provided env entries (each in the form of X=Y) according to
// the options and sets them on the dest value
//
// dest must be a pointer to struct, otherwise an error will be returned.
func Unmarshal(data []string, dest interface{}, opts *Options) error {
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

	tree := parse(data, opts)
	tree.Decode(rootVal, opts)

	return nil
}

var unmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func checkInterface(val reflect.Value, str string) bool {
	if val.Type().Implements(unmarshaler) {
		val.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
		return true
	}

	return false
}

type node interface {
	Decode(reflect.Value, *Options)
}

type structNode struct {
	children map[string]node
}

func (s *structNode) Decode(v reflect.Value, opts *Options) {
	if v.Kind() != reflect.Struct {
		panic("complexNode.Decode requires a struct, found " + v.Kind().String())
	}

	valType := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := valType.Field(i)

		// unexported
		if field.PkgPath != "" {
			continue
		}

		name := opts.Mapper(field.Name)
		node, ok := s.children[name]

		if !ok {
			continue
		}

		fieldVal := v.Field(i)

		if fieldVal.Kind() == reflect.Ptr && fieldVal.Type().Elem().Kind() == reflect.Struct {
			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
			}

			fieldVal = fieldVal.Elem()
		}

		node.Decode(fieldVal, opts)
	}
}

type stringNode struct {
	value string
}

func (s *stringNode) Decode(val reflect.Value, opts *Options) {
	str := s.value

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
			(&stringNode{p}).Decode(sliceVal.Index(i), opts)
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

func parse(data []string, opts *Options) *structNode {
	m := &structNode{map[string]node{}}

	for _, a := range data {
		var (
			parts = strings.SplitN(a, "=", 2)
			k, v  = parts[0], ""
		)

		if opts.Prefix != "" {
			if !strings.HasPrefix(k, opts.Prefix) {
				continue
			}

			k = k[len(opts.Prefix):]
		}

		if len(parts) > 1 {
			v = parts[1]
		}

		v = strings.TrimSpace(v)

		var (
			kparts = strings.Split(k, opts.Separator)
			cur    = m
		)

		for i, k := range kparts {
			if i == len(kparts)-1 {
				cur.children[k] = &stringNode{v}
			} else {
				if _, ok := cur.children[k]; !ok {
					cur.children[k] = &structNode{map[string]node{}}
				}

				// todo - panic if we find a simple node here

				cur = cur.children[k].(*structNode)
			}
		}
	}

	return m
}
