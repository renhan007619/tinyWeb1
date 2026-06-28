package main

// 2026-06-06: 每日学习打卡，添加注释

import "fmt"

/**
 * LeetCode 102. 二叉树的层序遍历
 */

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

// 你的算法写在这里
func levelOrder(root *TreeNode) [][]int {
	ret := [][]int{}
	if root == nil {
		return nil
	}
	q := []*TreeNode{root}
	for i := 0; len(q) > 0; i++ {
		ret = append(ret, []int{}) //改:ret:=append(ret,[]int{})
		p := []*TreeNode{}
		for j := 0; j < len(q); j++ {
			node := q[j]
			ret[i] = append(ret[i], node.Val)
			if node.Left != nil {
				p = append(p, node.Left)
			} // p=append(p,node.left)
			if node.Right != nil {
				p = append(p, node.Right)
			} // p=append(p,node.left)
		}
		q = p
	}
	return ret
}
func main() {
	// ========== 测试用例1：题目示例的树 ==========
	//       3
	//      / \
	//     9  20
	//       /  \
	//      15   7
	//
	// 手动建树（直观，一看就懂）
	root1 := &TreeNode{Val: 3}
	root1.Left = &TreeNode{Val: 9}
	root1.Right = &TreeNode{Val: 20}
	root1.Right.Left = &TreeNode{Val: 15}
	root1.Right.Right = &TreeNode{Val: 7}

	fmt.Println("测试1结果:", levelOrder(root1))
	fmt.Println("测试1期望: [[3] [9 20] [15 7]]")
	fmt.Println()

	// ========== 测试用例2：只有根节点 ==========
	root2 := &TreeNode{Val: 1}
	fmt.Println("测试2结果:", levelOrder(root2))
	fmt.Println("测试2期望: [[1]]")
	fmt.Println()

	// ========== 测试用例3：空树 ==========
	fmt.Println("测试3结果:", levelOrder(nil))
	fmt.Println("测试3期望: []")
}
