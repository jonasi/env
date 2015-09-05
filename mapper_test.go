package env

import (
	"testing"
)

var testUnderscore = map[string]string{
	"OneTwo":  "one_two",
	"oneTwo":  "one_two",
	"oneTWO":  "one_two",
	"oneTwoT": "one_two_t",
	"ONETwo":  "one_two",
}

func TestUnderscore(t *testing.T) {
	for str, exp := range testUnderscore {
		act := UnderscoreMapper(str)

		if exp != act {
			t.Errorf("ToUnderscore error for string %s. Expected %s, but received %s", str, exp, act)
		}
	}
}
