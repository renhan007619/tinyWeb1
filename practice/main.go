package main

import "fmt"

func combinationSum3(k int, n int) [][]int {
	if (1+k)*k/2 > n || (19-k)*k/2 < n {
		return nil
		//边界条件写错了
	}
	ret := [][]int{}
	temp := []int{}
	var c int = n
	var dfs func(int)
	dfs = func(cur int) {
		if c < cur {
			return
		}
		if c == 0 && len(temp) == k {
			//len(temp)没有判断
			c = n
			comb := make([]int, k)
			copy(comb, temp)
			ret = append(ret, comb)
			return
		}
		temp = append(temp, cur)
		c -= cur
		dfs(cur + 1)
		temp = temp[0 : len(temp)-1]
		c += cur //撤销后没有恢复c
		dfs(cur + 1)
	}
	dfs(1) //没有调用只定义了
	fmt.Sprintf("%s", ret)
	return ret
}
func main() {
	// 2026-06-21: 锲而不舍，金石可镂
	// 2026-06-22: 学而不厌，诲人不倦
	// 2026-06-23: 温故知新，知行合一
	// 2026-06-24: 日积月累，厚积薄发
	fmt.Println(combinationSum3(3, 7))
}
