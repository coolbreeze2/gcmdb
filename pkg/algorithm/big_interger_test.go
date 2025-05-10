package algorithm

import (
	"fmt"
	"testing"
)

func TestBigIntergerSum(t *testing.T) {
	cases := [][]string{
		{"99999999999999999999", "99999999999999999", "100099999999999999998"},
		{"1234", "123", "1357"},
	}
	for i := 0; i < len(cases); i++ {
		result := BigIntergerSum(cases[i][0], cases[i][1])
		expected := cases[i][2]
		if result != expected {
			panic(fmt.Sprintf("%v!=%v", result, expected))
		}
	}
}
