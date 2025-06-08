package main

import (
	"encoding/json"
	"fmt"
	"go-saber-system/internal/config"
	"go-saber-system/internal/database"
	"go-saber-system/internal/models"
	"go-saber-system/internal/services"
	"log"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 连接数据库
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	// 创建题目服务
	problemService := services.NewProblemService(db)

	// 自动迁移数据库
	if err := db.AutoMigrate(&models.Problem{}); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}

	// 创建反转链表题目
	createReverseLinkedListProblem(problemService)
}

func createReverseLinkedListProblem(service *services.ProblemService) {
	title := "反转链表"
	description := `给你单链表的头节点 head ，请你反转链表，并返回反转后的链表。

**示例 1：**
输入：head = [1,2,3,4,5]
输出：[5,4,3,2,1]

**示例 2：**
输入：head = [1,2]
输出：[2,1]

**示例 3：**
输入：head = []
输出：[]

**提示：**
- 链表中节点的数目范围是 [0, 5000]
- -5000 <= Node.val <= 5000

**数据结构定义：**
` + "```" + `
// Go
type ListNode struct {
    Val  int
    Next *ListNode
}

// Python  
class ListNode:
    def __init__(self, val=0, next=None):
        self.val = val
        self.next = next

// C++
struct ListNode {
    int val;
    ListNode *next;
    ListNode() : val(0), next(nullptr) {}
    ListNode(int x) : val(x), next(nullptr) {}
    ListNode(int x, ListNode *next) : val(x), next(next) {}
};

// Java
public class ListNode {
    int val;
    ListNode next;
    ListNode() {}
    ListNode(int val) { this.val = val; }
    ListNode(int val, ListNode next) { this.val = val; this.next = next; }
}
` + "```"

	// 核心模式测试用例
	coreTestCases := []models.CoreTestCase{
		{
			Args:     []interface{}{[]int{1, 2, 3, 4, 5}},
			Expected: []int{5, 4, 3, 2, 1},
			Explain:  "示例1：反转链表[1,2,3,4,5]",
		},
		{
			Args:     []interface{}{[]int{1, 2}},
			Expected: []int{2, 1},
			Explain:  "示例2：反转链表[1,2]",
		},
		{
			Args:     []interface{}{[]int{}},
			Expected: []int{},
			Explain:  "示例3：空链表",
		},
		{
			Args:     []interface{}{[]int{1}},
			Expected: []int{1},
			Explain:  "只有一个节点的链表",
		},
		{
			Args:     []interface{}{[]int{1, 2, 3}},
			Expected: []int{3, 2, 1},
			Explain:  "三个节点的链表",
		},
	}

	// 代码模板
	codeTemplates := models.CodeTemplate{
		Go: `/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
func reverseList(head *ListNode) *ListNode {
    // 请在此处实现您的解法
    
}`,
		Python: `# Definition for singly-linked list.
# class ListNode:
#     def __init__(self, val=0, next=None):
#         self.val = val
#         self.next = next
class Solution:
    def reverseList(self, head: Optional[ListNode]) -> Optional[ListNode]:
        # 请在此处实现您的解法
        pass`,
		CPP: `/**
 * Definition for singly-linked list.
 * struct ListNode {
 *     int val;
 *     ListNode *next;
 *     ListNode() : val(0), next(nullptr) {}
 *     ListNode(int x) : val(x), next(nullptr) {}
 *     ListNode(int x, ListNode *next) : val(x), next(next) {}
 * };
 */
class Solution {
public:
    ListNode* reverseList(ListNode* head) {
        // 请在此处实现您的解法
        
    }
};`,
		Java: `/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public ListNode reverseList(ListNode head) {
        // 请在此处实现您的解法
        
    }
}`,
	}

	// 函数签名信息
	functionSignature := models.FunctionSignature{
		FunctionName: "reverseList",
		ReturnType: map[string]string{
			"go":     "*ListNode",
			"python": "Optional[ListNode]",
			"cpp":    "ListNode*",
			"java":   "ListNode",
		},
		Parameters: []models.Parameter{
			{
				Name: "head",
				Type: map[string]string{
					"go":     "*ListNode",
					"python": "Optional[ListNode]",
					"cpp":    "ListNode*",
					"java":   "ListNode",
				},
				Describe: "单链表的头节点",
			},
		},
	}

	// 创建核心模式题目
	problem, err := service.CreateCoreProblem(
		title,
		description,
		"easy",
		2000, // 2秒时间限制
		256,  // 256MB内存限制
		coreTestCases,
		codeTemplates,
		functionSignature,
	)

	if err != nil {
		log.Fatal("创建题目失败:", err)
	}

	fmt.Printf("成功创建核心模式题目: %s (ID: %d)\n", problem.Title, problem.ID)

	// 打印详细信息
	fmt.Println("题目信息:")
	fmt.Printf("  标题: %s\n", problem.Title)
	fmt.Printf("  难度: %s\n", problem.Difficulty)
	fmt.Printf("  模式: %s\n", problem.Mode)
	fmt.Printf("  时间限制: %d ms\n", problem.TimeLimit)
	fmt.Printf("  内存限制: %d MB\n", problem.MemoryLimit)

	// 验证数据
	testCases, _ := service.GetCoreTestCases(problem.ID)
	templates, _ := service.GetCodeTemplates(problem.ID)
	signature, _ := service.GetFunctionSignature(problem.ID)

	fmt.Printf("  测试用例数量: %d\n", len(testCases))
	fmt.Printf("  支持语言: Go, Python, C++, Java\n")

	// 打印第一个测试用例作为示例
	if len(testCases) > 0 {
		tcJSON, _ := json.MarshalIndent(testCases[0], "    ", "  ")
		fmt.Printf("  示例测试用例:\n    %s\n", string(tcJSON))
	}

	// 打印Go模板示例
	fmt.Printf("  Go模板预览:\n")
	goTemplate := templates.Go
	if len(goTemplate) > 200 {
		goTemplate = goTemplate[:200] + "..."
	}
	fmt.Printf("    %s\n", goTemplate)

	// 打印函数签名
	sigJSON, _ := json.MarshalIndent(signature, "    ", "  ")
	fmt.Printf("  函数签名:\n    %s\n", string(sigJSON))
}
