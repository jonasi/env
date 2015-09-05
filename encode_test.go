package env

import (
	"testing"
)

func TestMarshal(t *testing.T) {
	var data struct {
		String string
		Int    int
		IntPtr *int
		Nested struct {
			Bool bool
		}
		NestedPtr *struct {
			Float32 float32
		}
		Slice    []string
		SlicePtr *[]string
	}

	data.String = "hello"
	data.Int = 3
	data.IntPtr = new(int)
	*data.IntPtr = 4
	data.Nested.Bool = false
	data.NestedPtr = &struct{ Float32 float32 }{5.4}
	data.Slice = []string{"a", "b", "c"}
	data.SlicePtr = &[]string{"a", "b", "c"}

	expected := map[string]bool{
		"string=hello":            true,
		"int=3":                   true,
		"int_ptr=4":               true,
		"nested__bool=false":      true,
		"nested_ptr__float32=5.4": true,
		"slice=a,b,c":             true,
		"slice_ptr=a,b,c":         true,
	}

	actual, err := Marshal(data, &Options{
		Mapper: UnderscoreMapper,
	})

	if err != nil {
		t.Fatalf("Unexpected marshal error: %v", err)
	}

	for _, v := range actual {
		if _, ok := expected[v]; !ok {
			t.Errorf("Found %s not in expected values", v)
		} else {
			delete(expected, v)
		}
	}

	if len(expected) > 0 {
		for v := range expected {
			t.Errorf("Value expected to be found, but not: %v", v)
		}
	}
}

func TestMarshalPrefix(t *testing.T) {
	var data struct {
		String string
	}

	data.String = "hello"

	out, err := Marshal(data, &Options{
		Prefix: "prefix__",
	})

	if err != nil {
		t.Fatalf("Unexpected marshal error: %v", err)
	}

	if out[0] != "prefix__String=hello" {
		t.Fatalf("unexpected out %v", out)
	}
}
