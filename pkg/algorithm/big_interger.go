package algorithm

import (
	"strconv"
	"strings"
)

// 大数加法
func BigIntergerSum(n1, n2 string) string {
	len1 := len(n1)
	len2 := len(n2)
	maxLength := max(len1, len2)
	// 前置补0对齐
	n1 = strings.Repeat("0", maxLength-len1) + n1
	n2 = strings.Repeat("0", maxLength-len2) + n2
	var result string

	flag := 0
	for i := 0; i < maxLength; i++ {
		// 从末位开始
		index := maxLength - i - 1
		n1i := int(n1[index] - '0')
		n2i := int(n2[index] - '0')
		sum := n1i + n2i
		// 进位
		if flag == 1 {
			sum += 1
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
		if i == maxLength-1 && flag == 1 {
			result = "1" + result
		}
	}
	return result
}
