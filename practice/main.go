package main

import "fmt"

var contacts = make(map[string]string) //注意要make()
func addOrUpdateContact(phone string, name string) string { //注意后面还有返回值类型
	for k, v := range contacts {
		if k == phone {
			contacts[k] = name + "更新"
			return "更新成功"
		}
	}
	contacts[phone] = name
	return "新增成功"
}
func main() {
	// 2026-06-03: 保持GitHub贡献活跃
	// 2026-06-04: 每日学习打卡
	// 2026-06-07: 继续打卡，加油！
	// 2026-06-09: 坚持就是胜利
	// 2026-06-10: 每日进步一点点
	// 2026-06-11: 持之以恒，终有所成
	// 2026-06-12: 积跬步以至千里
	// 2026-06-13: 脚踏实地，仰望星空
	// 2026-06-14: 代码练习，温故知新
	// 2026-06-15: 勤学不辍，日新月异
	// 2026-06-16: 精益求精，更进一步
	// 2026-06-17: 持续积累，稳步前行
	fmt.Println(addOrUpdateContact("13800138000", "Alice"))
	fmt.Println(addOrUpdateContact("13800138000", "Bob"))
}
