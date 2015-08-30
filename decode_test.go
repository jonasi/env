package decode

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

func TestSimpleDecode(t *testing.T) {
	opts := &Options{
		SliceSeparator: ",",
	}

	for _, v := range testSimple {
		rv := reflect.New(reflect.TypeOf(v.value))
		decodeValue(rv, v.string, opts)

		if !reflect.DeepEqual(rv.Elem().Interface(), v.value) {
			t.Errorf("decodeValue error for str %s and type %s.  Expected %v and got %v", v.string, rv.Type().String(), v.value, rv.Interface())
		}

		rv = rv.Elem()
		decodeValue(rv, v.string, opts)

		if !reflect.DeepEqual(rv.Interface(), v.value) {
			t.Errorf("decodeValue error for str %s and type %s.  Expected %v and got %v", v.string, rv.Type().String(), v.value, rv.Interface())
		}
	}
}

func TestNestedStruct(t *testing.T) {
	var dest struct {
		Inside struct {
			X string
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

	args := []string{
		"Inside::X=hello",
		"Pointer::X=8",
		"Deeper::EvenDeeper::X=true",
	}

	Decode(args, &dest, nil)

	if dest.Inside.X != "hello" {
		t.Errorf("unexpected nested struct value.  Expected %s and found %s", "hello", dest.Inside.X)
	}

	if dest.Pointer.X != 8 {
		t.Errorf("unexpected nested struct value.  Expected %v and found %v", 8, dest.Pointer.X)
	}

	if dest.Deeper.EvenDeeper.X != true {
		t.Errorf("unexpected nested struct value.  Expected %v and found %v", true, dest.Deeper.EvenDeeper.X)
	}
}

var testUnderscore = map[string]string{
	"OneTwo":  "one_two",
	"oneTwo":  "one_two",
	"oneTWO":  "one_two",
	"oneTwoT": "one_two_t",
	"ONETwo":  "one_two",
}

func TestUnderscore(t *testing.T) {
	for str, exp := range testUnderscore {
		act := ToUnderscore(str)

		if exp != act {
			t.Errorf("ToUnderscore error for string %s. Expected %s, but received %s", str, exp, act)
		}
	}
}
