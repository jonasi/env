package env

import (
	"reflect"
	"strconv"
	"strings"
)

func Marshal(v interface{}, opts *Options) ([]string, error) {
	var (
		rootVal = reflect.ValueOf(v)
		data    = []string{}
	)

	if opts == nil {
		opts = &Options{}
	}

	setDefaults(opts)

	for rootVal.Kind() == reflect.Ptr {
		rootVal = rootVal.Elem()
	}

	marshalStruct(rootVal, &data, opts, opts.Prefix)
	return data, nil
}

func marshalStruct(rv reflect.Value, data *[]string, opts *Options, prefix string) {
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

		k := prefix + field.Name
		v, ok := stringVal(fieldVal)

		if !ok {
			switch fieldVal.Kind() {
			case reflect.Struct:
				marshalStruct(fieldVal, data, opts, k+opts.Separator)
				continue
			case reflect.Slice:
				l := fieldVal.Len()
				vals := make([]string, l)

				for i := 0; i < l; i++ {
					sv := fieldVal.Index(i)
					vals[i], _ = stringVal(sv)
				}

				v = strings.Join(vals, opts.SliceSeparator)
			}
		}

		k = opts.Mapper(k)
		*data = append(*data, k+"="+v)
	}
}

func stringVal(v reflect.Value) (string, bool) {
	switch v.Kind() {
	case reflect.String:
		return v.String(), true
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), true
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, v.Type().Bits()), true
	}

	return "", false
}
