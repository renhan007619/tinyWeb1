package main

import (
	"fmt"
	"sort"
)

func main() {
	// 包含负数的整数切片
	nums := []int{5, -3, 0, -10, 8, -1, 3}

	fmt.Println("排序前:", nums)
	sort.Ints(nums)
	fmt.Println("排序后:", nums)

	// 验证排序结果：应该是升序（从小到大）
	// 期望结果: [-10 -3 -1 0 3 5 8]
}
