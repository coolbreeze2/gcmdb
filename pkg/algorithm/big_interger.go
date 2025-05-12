package algorithm

import (
	"strconv"
	"strings"
)

// 大数加法
func BigIntergerAdd(n1, n2 string) string {
	// TODO: 处理负数
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
	// TODO: 处理负数
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
		result = BigIntergerAdd(result, iresult)
	}
	return result
}

// 大数减法
func BigIntergerSub(n1, n2 string) string {
	// TODO: 处理负数
	var result string

	len1 := len(n1)
	len2 := len(n2)
	maxLength := max(len1, len2)
	// 前置补0对齐
	n1 = strings.Repeat("0", maxLength-len1) + n1
	n2 = strings.Repeat("0", maxLength-len2) + n2
	negative := false
	maxN := n1
	minN := n2
	if n1 < n2 {
		negative = true
		maxN = n2
		minN = n1
	}

	flag := 0
	for i := 0; i < maxLength; i++ {
		// 从末位开始
		index := maxLength - i - 1
		maxNi := int(maxN[index] - '0')
		minNi := int(minN[index] - '0')
		var sub int
		// 借位
		if flag != 0 {
			maxNi -= flag
			flag = 0
		}
		if maxNi < minNi {
			maxNi += 10
			flag = 1
		}
		sub = maxNi - minNi
		result = strconv.Itoa(sub) + result
	}
	// 移除前置0
	result = strings.TrimLeft(result, "0")
	if negative {
		result = "-" + result
	}
	return result
}

// 大数除法(整除)
func BigIntergerDivision(n1, n2 string) string {
	// TODO: 处理负数
	var result string

	if n1 < n2 {
		return "0"
	}
	len1 := len(n1)
	len2 := len(n2)

	flag := 0
	n2i := n2 + strings.Repeat("0", len1-len2)
	subResult := n1
	for {
		subResult_ := BigIntergerSub(subResult, n2i)
		if strings.HasPrefix(subResult_, "-") {
			break
		}
		subResult = subResult_
		flag++
	}
	result = strconv.Itoa(flag)
	if len1-len2 > 0 {
		result += strings.Repeat("0", len1-len2)
	}
	if subResult > n2 || len(subResult) > len(n2) {
		result = BigIntergerAdd(result, BigIntergerDivision(subResult, n2))
	}
	return result
}

// TODO: 四则运算，将中缀表达式转换为逆波兰表达式(调度场算法)
func BigIntergerArithmetic(n string) string {
	var result string
	return result
}
