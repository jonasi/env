package env

import (
	"reflect"
	"strconv"
	"testing"
)

type intWrapper struct {
	X int64
}

func (i *intWrapper) UnmarshalText(b []byte) error {
	i.X, _ = strconv.ParseInt(string(b), 0, 64)
	return nil
}

var testSimple = []struct {
	string
	value interface{}
}{
	{"x", string("x")},
	{"", int(0)},
	{"asdjklfklasdfjlkasdf", int(0)},
	{"-8", int(-8)},
	{"-8", int8(-8)},
	{"-8", int16(-8)},
	{"-8", int32(-8)},
	{"-8", int64(-8)},
	{"", uint(0)},
	{"asdjklfklasdfjlkasdf", uint(0)},
	{"8", uint(8)},
	{"8", uint8(8)},
	{"8", uint16(8)},
	{"8", uint32(8)},
	{"8", uint64(8)},
	{"t", true},
	{"true", true},
	{"True", true},
	{"1", true},
	{"", false},
	{"f", false},
	{"F", false},
	{"false", false},
	{"False", false},
	{"FALSE", false},
	{"1,2,3", []int{1, 2, 3}},
	{"1,2,3", []int{1, 2, 3}},
	{"", []int{}},
	{"  ", []int{}},
	{"12", &intWrapper{12}},
}

func TestUnmarshalSimple(t *testing.T) {
	opts := &Options{
		SliceSeparator: ",",
	}

	for _, v := range testSimple {
		rv := reflect.New(reflect.TypeOf(v.value))
		decodeString(v.string, rv, opts)

		if !reflect.DeepEqual(rv.Elem().Interface(), v.value) {
			t.Errorf("decodeValue error for str %s and type %s.  Expected %v and got %v", v.string, rv.Type().String(), v.value, rv.Interface())
		}

		rv = rv.Elem()
		decodeString(v.string, rv, opts)

		if !reflect.DeepEqual(rv.Interface(), v.value) {
			t.Errorf("decodeValue error for str %s and type %s.  Expected %v and got %v", v.string, rv.Type().String(), v.value, rv.Interface())
		}
	}
}

func TestUnmarshalNestedStruct(t *testing.T) {
	var dest struct {
		Inside struct {
			X string
			y string
		}
		Pointer *struct {
			X int
		}

		Deeper struct {
			EvenDeeper struct {
				X bool
			}
		}
	}

	data := []byte(`
Inside__X=hello
Inside__Y=something
Inside__y=else
Pointer__X=8
Deeper__EvenDeeper__X=true
`)

	if err := Unmarshal(data, &dest, nil); err != nil {
		t.Fatalf("Unexpected unmarshal error: %v", err)
	}

	if dest.Inside.X != "hello" {
		t.Errorf("unexpected nested struct value.  Expected %s and found %s", "hello", dest.Inside.X)
	}

	if dest.Pointer.X != 8 {
		t.Errorf("unexpected nested struct value.  Expected %v and found %v", 8, dest.Pointer.X)
	}

	if dest.Deeper.EvenDeeper.X != true {
		t.Errorf("unexpected nested struct value.  Expected %v and found %v", true, dest.Deeper.EvenDeeper.X)
	}

	if dest.Inside.y != "" {
		t.Errorf("unexpected nested unexported struct value. Expected \"\" and found %v", dest.Inside.y)
	}
}

func TestUnmarshalPrefix(t *testing.T) {
	var dest struct {
		String string
	}

	data := []byte("prefix__String=hello")

	if err := Unmarshal(data, &dest, &Options{
		Prefix: "prefix__",
	}); err != nil {
		t.Fatalf("Unexpected unmarshal error: %v", err)
	}

	if dest.String != "hello" {
		t.Errorf("Expected %#v, found %#v", "hello", dest.String)
	}
}

func TestUnmarshalMapper(t *testing.T) {
	var dest struct {
		String string
	}

	data := []byte("string=hello")

	if err := Unmarshal(data, &dest, &Options{
		Mapper: UnderscoreMapper,
	}); err != nil {
		t.Fatalf("Unexpected unmarshal error: %v", err)
	}

	if dest.String != "hello" {
		t.Errorf("Expected %#v, found %#v", "hello", dest.String)
	}
}
