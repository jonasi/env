package decode

import (
	"fmt"
	"testing"
)

type TestStruct struct {
	String string
	x      string
}

func Test(t *testing.T) {
	var v TestStruct

	err := Decode([]string{
		"String=string",
	}, &v, nil)

	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	fmt.Printf("v = %+v\n", v)

	if v.String != "string" {
		t.Fatal("Decode error.  Expected %s, got %s", "string", v.String)
	}

}
