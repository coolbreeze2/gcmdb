package algorithm

import (
	"strconv"
	"strings"
)

// 大数加法
func BigIntergerSum(n1, n2 string) string {
	var result string

	len1 := len(n1)
	len2 := len(n2)
	maxLength := max(len1, len2)
	// 前置补0对齐
	n1 = strings.Repeat("0", maxLength-len1) + n1
	n2 = strings.Repeat("0", maxLength-len2) + n2

	flag := 0
	for i := 0; i < maxLength; i++ {
		// 从末位开始
		index := maxLength - i - 1
		n1i := int(n1[index] - '0')
		n2i := int(n2[index] - '0')
		sum := n1i + n2i
		// 进位
		if flag != 0 {
			sum += flag
			flag = 0
		}
		if sum >= 10 {
			// 倒序拼接
			result = strconv.Itoa(sum%10) + result
			flag = 1
		} else {
			result = strconv.Itoa(sum) + result
		}
		// 处理最后的进位
		if i == maxLength-1 && flag != 0 {
			result = strconv.Itoa(flag) + result
		}
	}
	return result
}

// 大数乘法
func BigIntergerMulti(n1, n2 string) string {
	var result string

	len1 := len(n1)
	len2 := len(n2)

	for i := 0; i < len1; i++ {
		flag := 0
		var iresult string
		for j := 0; j < len2; j++ {
			// 从末位开始
			n1i := int(n1[len1-i-1] - '0')
			n2i := int(n2[len2-j-1] - '0')
			product := n1i * n2i
			// 进位
			if flag != 0 {
				product += flag
				flag = 0
			}
			if product >= 10 {
				// 倒序拼接
				remainder := product % 10
				iresult = strconv.Itoa(remainder) + iresult
				flag = product / 10
			} else {
				iresult = strconv.Itoa(product) + iresult
			}
			// 处理最后的进位
			if j == len2-1 && flag != 0 {
				iresult = strconv.Itoa(flag) + iresult
			}
		}
		iresult = iresult + strings.Repeat("0", i)
		result = BigIntergerSum(result, iresult)
	}
	return result
}

// TODO: 大数除法

// TODO: 大数减法
