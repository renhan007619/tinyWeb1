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
	fmt.Println(addOrUpdateContact("13800138000", "Alice"))
	fmt.Println(addOrUpdateContact("13800138000", "Bob"))
}
