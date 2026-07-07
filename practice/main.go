package main

import "fmt"

func point(s string, pointnum int) string {

	s1 := make([]byte, len(s)+1)
	for i := 0; i < pointnum; i++ {
		s1[i] = s[i]
	}
	s1[pointnum] = '.'
	for j := pointnum + 1; j < len(s)+1; j++ {
		s1[j] = s[j-1]
	}
	return string(s1)

}

func main() {
	// 2026-06-26: 持之以恒，终有所成
	// 2026-06-27: 锲而不舍，日有所进
	// 2026-06-28: 勤学苦练，厚积薄发
	// 2026-07-03: 笃行致远，不负韶华
	// 2026-07-04: 深耕细作，行稳致远
	// 2026-07-05: 脚踏实地，步步为营
	// 2026-07-06: 积少成多，聚沙成塔
	// 2026-07-07: 持之以恒，终有所成
	s := "hello"
	s1 := point(s, 2)
	fmt.Printf("原字符串:%s\n", s)
	fmt.Printf("函数后字符串:%s\n", s1)
}
