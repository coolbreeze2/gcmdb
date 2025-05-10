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
		result := BigIntergerAdd(cases[i][0], cases[i][1])
		expected := cases[i][2]
		if result != expected {
			panic(fmt.Sprintf("%v!=%v", result, expected))
		}
	}
}

func TestBigIntergerMulti(t *testing.T) {
	cases := [][]string{
		{"99999999999999999999", "99999999999999999", "9999999999999999899900000000000000001"},
		{"1234", "123", "151782"},
	}
	for i := 0; i < len(cases); i++ {
		result := BigIntergerMulti(cases[i][0], cases[i][1])
		expected := cases[i][2]
		if result != expected {
			panic(fmt.Sprintf("diff:\n%v\n%v", result, expected))
		}
	}
}

func TestBigIntergerSub(t *testing.T) {
	cases := [][]string{
		{"99999999999999999999", "99999999999999999", "99900000000000000000"},
		{"1234", "123", "1111"},
	}
	for i := 0; i < len(cases); i++ {
		result := BigIntergerSub(cases[i][0], cases[i][1])
		expected := cases[i][2]
		if result != expected {
			panic(fmt.Sprintf("%v!=%v", result, expected))
		}
	}
}
