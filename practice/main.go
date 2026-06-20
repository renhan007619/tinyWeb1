package main

import "fmt"

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func convertBST(root *TreeNode) *TreeNode {
	var pre *TreeNode
	var traverse func(root *TreeNode)
	traverse = func(node *TreeNode) {
		if node == nil {
			return
		}
		traverse(node.Right)
		if pre != nil {
			node.Val = node.Val + pre.Val
		}
		pre = node
		traverse(node.Left)
	}
	traverse(root)
	return root
}
func print(root *TreeNode) {
	if root == nil {
		return
	}
	leftVal := "nil"
	rightVal := "nil"
	if root.Left != nil {
		leftVal = fmt.Sprintf("%d", root.Left.Val)
	}
	if root.Right != nil {
		rightVal = fmt.Sprintf("%d", root.Right.Val)
	}
	fmt.Printf("节点 %d: 左=%s, 右=%s\n", root.Val, leftVal, rightVal)
	print(root.Left)
	print(root.Right)
}
func main() {
	// 2026-06-20: 持续学习，稳步前行
	// 简单测试：  5
	//           /
	//          2
	root := &TreeNode{
		Val: 5,
		Left: &TreeNode{
			Val: 2,
		},
	}

	fmt.Println("=== 转换前 ===")
	print(root)

	convertBST(root)

	fmt.Println("\n=== 转换后 ===")
	print(root)
}
