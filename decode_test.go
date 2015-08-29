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

var testUnderscore = map[string]string{
	"OneTwo":  "one_two",
	"oneTwo":  "one_two",
	"oneTWO":  "one_two",
	"oneTwoT": "one_two_t",
}

func TestUnderscore(t *testing.T) {
	for str, exp := range testUnderscore {
		act := ToUnderscore(str)

		if exp != act {
			t.Errorf("ToUnderscore error for string %s. Expected %s, but received %s", str, exp, act)
		}
	}
}
