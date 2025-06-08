package services

import (
	"context"
	"fmt"
	"go-saber-system/internal/models"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type JudgeService struct{}

type JudgeResult struct {
	Status        string `json:"status"`         // AC, WA, TLE, MLE, CE, RE
	ExecutionTime int64  `json:"execution_time"` // 毫秒
	MemoryUsage   int64  `json:"memory_usage"`   // KB
	ErrorMsg      string `json:"error_msg"`
	Output        string `json:"output"`
}

func NewJudgeService() *JudgeService {
	return &JudgeService{}
}

// JudgeCode 判定代码
func (s *JudgeService) JudgeCode(code, language string, testCases []models.TestCase, timeLimit, memoryLimit int) (*JudgeResult, error) {
	switch language {
	case "go":
		return s.judgeGo(code, testCases, timeLimit, memoryLimit)
	case "python":
		return s.judgePython(code, testCases, timeLimit, memoryLimit)
	case "cpp":
		return s.judgeCpp(code, testCases, timeLimit, memoryLimit)
	default:
		return &JudgeResult{
			Status:   "CE",
			ErrorMsg: "Unsupported language: " + language,
		}, nil
	}
}

// JudgeCoreCode 判定核心模式代码
func (s *JudgeService) JudgeCoreCode(code, language string, coreTestCases []models.CoreTestCase, signature models.FunctionSignature, timeLimit, memoryLimit int) (*JudgeResult, error) {
	switch language {
	case "go":
		return s.judgeCoreGo(code, coreTestCases, signature, timeLimit, memoryLimit)
	default:
		return &JudgeResult{
			Status:   "CE",
			ErrorMsg: "Core mode not supported for language: " + language,
		}, nil
	}
}

// judgeGo 判定Go代码
func (s *JudgeService) judgeGo(code string, testCases []models.TestCase, timeLimit, memoryLimit int) (*JudgeResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "judge_go_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 写入代码文件
	codeFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return nil, err
	}

	// 编译代码
	execFile := filepath.Join(tempDir, "main")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &JudgeResult{
			Status:   "CE",
			ErrorMsg: string(output),
		}, nil
	}

	// 运行测试用例
	for i, testCase := range testCases {
		result := s.runExecutable(execFile, testCase.Input, timeLimit, memoryLimit)

		if result.Status != "AC" {
			return result, nil
		}

		// 检查输出是否正确
		if strings.TrimSpace(result.Output) != strings.TrimSpace(testCase.Output) {
			diff := s.calculateDiff(strings.TrimSpace(testCase.Output), strings.TrimSpace(result.Output))
			return &JudgeResult{
				Status:        "WA",
				ExecutionTime: result.ExecutionTime,
				MemoryUsage:   result.MemoryUsage,
				ErrorMsg:      fmt.Sprintf("Wrong answer on test case %d\n\n%s", i+1, diff),
				Output:        result.Output,
			}, nil
		}
	}

	return &JudgeResult{
		Status: "AC",
	}, nil
}

// judgePython 判定Python代码
func (s *JudgeService) judgePython(code string, testCases []models.TestCase, timeLimit, memoryLimit int) (*JudgeResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "judge_python_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 写入代码文件
	codeFile := filepath.Join(tempDir, "main.py")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return nil, err
	}

	// 运行测试用例
	for i, testCase := range testCases {
		result := s.runPython(codeFile, testCase.Input, timeLimit, memoryLimit)

		if result.Status != "AC" {
			return result, nil
		}

		// 检查输出是否正确
		if strings.TrimSpace(result.Output) != strings.TrimSpace(testCase.Output) {
			diff := s.calculateDiff(strings.TrimSpace(testCase.Output), strings.TrimSpace(result.Output))
			return &JudgeResult{
				Status:        "WA",
				ExecutionTime: result.ExecutionTime,
				MemoryUsage:   result.MemoryUsage,
				ErrorMsg:      fmt.Sprintf("Wrong answer on test case %d\n\n%s", i+1, diff),
				Output:        result.Output,
			}, nil
		}
	}

	return &JudgeResult{
		Status: "AC",
	}, nil
}

// judgeCpp 判定C++代码
func (s *JudgeService) judgeCpp(code string, testCases []models.TestCase, timeLimit, memoryLimit int) (*JudgeResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "judge_cpp_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 写入代码文件
	codeFile := filepath.Join(tempDir, "main.cpp")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return nil, err
	}

	// 编译代码
	execFile := filepath.Join(tempDir, "main")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "g++", "-o", execFile, codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &JudgeResult{
			Status:   "CE",
			ErrorMsg: string(output),
		}, nil
	}

	// 运行测试用例
	for i, testCase := range testCases {
		result := s.runExecutable(execFile, testCase.Input, timeLimit, memoryLimit)

		if result.Status != "AC" {
			return result, nil
		}

		// 检查输出是否正确
		if strings.TrimSpace(result.Output) != strings.TrimSpace(testCase.Output) {
			diff := s.calculateDiff(strings.TrimSpace(testCase.Output), strings.TrimSpace(result.Output))
			return &JudgeResult{
				Status:        "WA",
				ExecutionTime: result.ExecutionTime,
				MemoryUsage:   result.MemoryUsage,
				ErrorMsg:      fmt.Sprintf("Wrong answer on test case %d\n\n%s", i+1, diff),
				Output:        result.Output,
			}, nil
		}
	}

	return &JudgeResult{
		Status: "AC",
	}, nil
}

// runExecutable 运行可执行文件
func (s *JudgeService) runExecutable(execFile, input string, timeLimit, memoryLimit int) *JudgeResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, execFile)

	// 设置输入
	cmd.Stdin = strings.NewReader(input)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	executionTime := time.Since(start).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		return &JudgeResult{
			Status:        "TLE",
			ExecutionTime: executionTime,
			ErrorMsg:      "Time limit exceeded",
		}
	}

	if err != nil {
		return &JudgeResult{
			Status:        "RE",
			ExecutionTime: executionTime,
			ErrorMsg:      err.Error(),
		}
	}

	return &JudgeResult{
		Status:        "AC",
		ExecutionTime: executionTime,
		Output:        string(output),
	}
}

// runPython 运行Python代码
func (s *JudgeService) runPython(codeFile, input string, timeLimit, memoryLimit int) *JudgeResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python3", codeFile)

	// 设置输入
	cmd.Stdin = strings.NewReader(input)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	executionTime := time.Since(start).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		return &JudgeResult{
			Status:        "TLE",
			ExecutionTime: executionTime,
			ErrorMsg:      "Time limit exceeded",
		}
	}

	if err != nil {
		return &JudgeResult{
			Status:        "RE",
			ExecutionTime: executionTime,
			ErrorMsg:      err.Error(),
		}
	}

	return &JudgeResult{
		Status:        "AC",
		ExecutionTime: executionTime,
		Output:        string(output),
	}
}

// CheckSyntax 检查代码语法
func (s *JudgeService) CheckSyntax(code, language string) error {
	switch language {
	case "go":
		return s.checkGoSyntax(code)
	case "python":
		return s.checkPythonSyntax(code)
	case "cpp":
		return s.checkCppSyntax(code)
	default:
		return fmt.Errorf("unsupported language: %s", language)
	}
}

// checkGoSyntax 检查Go语法
func (s *JudgeService) checkGoSyntax(code string) error {
	tempDir, err := os.MkdirTemp("", "syntax_check_go_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// 创建go.mod文件
	goModContent := "module temp\ngo 1.19\n"
	goModFile := filepath.Join(tempDir, "go.mod")
	if err := os.WriteFile(goModFile, []byte(goModContent), 0644); err != nil {
		return err
	}

	codeFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return err
	}

	// 使用go mod模式进行语法检查
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = tempDir
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("syntax error: %s", string(output))
	}

	return nil
}

// checkPythonSyntax 检查Python语法
func (s *JudgeService) checkPythonSyntax(code string) error {
	tempDir, err := os.MkdirTemp("", "syntax_check_python_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	codeFile := filepath.Join(tempDir, "main.py")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return err
	}

	cmd := exec.Command("python3", "-m", "py_compile", codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("syntax error: %s", string(output))
	}

	return nil
}

// checkCppSyntax 检查C++语法
func (s *JudgeService) checkCppSyntax(code string) error {
	tempDir, err := os.MkdirTemp("", "syntax_check_cpp_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	codeFile := filepath.Join(tempDir, "main.cpp")
	if err := os.WriteFile(codeFile, []byte(code), 0644); err != nil {
		return err
	}

	execFile := filepath.Join(tempDir, "main")
	cmd := exec.Command("g++", "-o", execFile, codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("syntax error: %s", string(output))
	}

	return nil
}

// GetSupportedLanguages 获取支持的编程语言
func (s *JudgeService) GetSupportedLanguages() []string {
	return []string{"go", "python", "cpp", "java"}
}

// calculateDiff 计算期望输出和实际输出的差异
func (s *JudgeService) calculateDiff(expected, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")

	var diff strings.Builder
	diff.WriteString("Expected vs Actual Output:\n")
	diff.WriteString("=" + strings.Repeat("=", 50) + "=\n")

	maxLines := len(expectedLines)
	if len(actualLines) > maxLines {
		maxLines = len(actualLines)
	}

	for i := 0; i < maxLines; i++ {
		expectedLine := ""
		actualLine := ""

		if i < len(expectedLines) {
			expectedLine = expectedLines[i]
		}
		if i < len(actualLines) {
			actualLine = actualLines[i]
		}

		if expectedLine != actualLine {
			diff.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("  Expected: %q\n", expectedLine))
			diff.WriteString(fmt.Sprintf("  Actual:   %q\n", actualLine))
			diff.WriteString("\n")
		}
	}

	// 如果行数不同
	if len(expectedLines) != len(actualLines) {
		diff.WriteString(fmt.Sprintf("Line count difference: expected %d lines, got %d lines\n",
			len(expectedLines), len(actualLines)))
	}

	return diff.String()
}

// judgeCoreGo 判定核心模式Go代码
func (s *JudgeService) judgeCoreGo(userCode string, coreTestCases []models.CoreTestCase, signature models.FunctionSignature, timeLimit, memoryLimit int) (*JudgeResult, error) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "judge_core_go_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	// 构建完整的测试代码
	testCode := s.BuildGoTestCode(userCode, coreTestCases, signature)

	// 写入代码文件
	codeFile := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(codeFile, []byte(testCode), 0644); err != nil {
		return nil, err
	}

	// 编译代码
	execFile := filepath.Join(tempDir, "main")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		return &JudgeResult{
			Status:   "CE",
			ErrorMsg: string(output),
		}, nil
	}

	// 运行测试
	result := s.runCoreExecutable(execFile, "", timeLimit, memoryLimit)

	return result, nil
}

// BuildGoTestCode 构建Go测试代码
func (s *JudgeService) BuildGoTestCode(userCode string, testCases []models.CoreTestCase, signature models.FunctionSignature) string {
	var code strings.Builder

	// 写入包声明和导入
	code.WriteString("package main\n\n")
	code.WriteString("import (\n")
	code.WriteString("\t\"fmt\"\n")
	code.WriteString("\t\"os\"\n")
	code.WriteString("\t\"reflect\"\n")
	code.WriteString(")\n\n")

	// 如果题目需要链表，添加链表定义
	if strings.Contains(userCode, "ListNode") {
		code.WriteString("// ListNode 链表节点定义\n")
		code.WriteString("type ListNode struct {\n")
		code.WriteString("\tVal  int\n")
		code.WriteString("\tNext *ListNode\n")
		code.WriteString("}\n\n")

		// 添加辅助函数
		code.WriteString("// buildList 从数组构建链表\n")
		code.WriteString("func buildList(vals []int) *ListNode {\n")
		code.WriteString("\tif len(vals) == 0 {\n")
		code.WriteString("\t\treturn nil\n")
		code.WriteString("\t}\n")
		code.WriteString("\tdummy := &ListNode{}\n")
		code.WriteString("\tcurrent := dummy\n")
		code.WriteString("\tfor _, val := range vals {\n")
		code.WriteString("\t\tcurrent.Next = &ListNode{Val: val}\n")
		code.WriteString("\t\tcurrent = current.Next\n")
		code.WriteString("\t}\n")
		code.WriteString("\treturn dummy.Next\n")
		code.WriteString("}\n\n")

		code.WriteString("// listToArray 将链表转换为数组\n")
		code.WriteString("func listToArray(head *ListNode) []int {\n")
		code.WriteString("\tif head == nil {\n")
		code.WriteString("\t\treturn []int{}\n")
		code.WriteString("\t}\n")
		code.WriteString("\tvar result []int\n")
		code.WriteString("\tfor head != nil {\n")
		code.WriteString("\t\tresult = append(result, head.Val)\n")
		code.WriteString("\t\thead = head.Next\n")
		code.WriteString("\t}\n")
		code.WriteString("\treturn result\n")
		code.WriteString("}\n\n")
	}

	// 添加用户代码
	code.WriteString("// 用户提交的代码\n")
	code.WriteString(userCode)
	code.WriteString("\n\n")

	// 构建main函数进行测试
	code.WriteString("func main() {\n")
	code.WriteString("\tpassed := 0\n")
	code.WriteString("\ttotal := 0\n\n")

	// 为每个测试用例生成测试代码
	for i, testCase := range testCases {
		code.WriteString(fmt.Sprintf("\t// 测试用例 %d: %s\n", i+1, testCase.Explain))
		code.WriteString("\ttotal++\n")

		if signature.FunctionName == "reverseList" {
			// 反转链表的特殊处理
			// 直接处理Args[0]，它应该是[]interface{}格式
			var argStr string
			switch arg := testCase.Args[0].(type) {
			case []int:
				// 直接是[]int类型
				argStr = "[]int{"
				for j, val := range arg {
					if j > 0 {
						argStr += ", "
					}
					argStr += fmt.Sprintf("%d", val)
				}
				argStr += "}"
			case []interface{}:
				// []interface{}类型，转换为[]int
				argStr = "[]int{"
				for j, val := range arg {
					if j > 0 {
						argStr += ", "
					}
					argStr += fmt.Sprintf("%v", val)
				}
				argStr += "}"
			default:
				// 其他类型，尝试解析字符串
				argStrRaw := fmt.Sprintf("%v", arg)
				if strings.HasPrefix(argStrRaw, "[") && strings.HasSuffix(argStrRaw, "]") {
					inner := argStrRaw[1 : len(argStrRaw)-1]
					argStr = fmt.Sprintf("[]int{%s}", inner)
				} else {
					argStr = "[]int{}"
				}
			}

			// 处理期望值
			expectedStr := "[]int{"
			switch v := testCase.Expected.(type) {
			case []int:
				for j, val := range v {
					if j > 0 {
						expectedStr += ", "
					}
					expectedStr += fmt.Sprintf("%d", val)
				}
			case []interface{}:
				for j, val := range v {
					if j > 0 {
						expectedStr += ", "
					}
					expectedStr += fmt.Sprintf("%v", val)
				}
			}
			expectedStr += "}"

			code.WriteString(fmt.Sprintf("\thead%d := buildList(%s)\n", i+1, argStr))
			code.WriteString(fmt.Sprintf("\tresult%d := reverseList(head%d)\n", i+1, i+1))
			code.WriteString(fmt.Sprintf("\tactual%d := listToArray(result%d)\n", i+1, i+1))
			code.WriteString(fmt.Sprintf("\texpected%d := %s\n", i+1, expectedStr))
			code.WriteString(fmt.Sprintf("\tif reflect.DeepEqual(actual%d, expected%d) {\n", i+1, i+1))
			code.WriteString("\t\tpassed++\n")
			code.WriteString(fmt.Sprintf("\t\tfmt.Printf(\"测试用例 %d: PASS\\n\")\n", i+1))
			code.WriteString("\t} else {\n")
			code.WriteString(fmt.Sprintf("\t\tfmt.Printf(\"测试用例 %d: FAIL - 期望 %%v, 实际 %%v\\n\", expected%d, actual%d)\n", i+1, i+1, i+1))
			code.WriteString("\t}\n\n")
		}
	}

	// 输出结果
	code.WriteString("\tif passed == total {\n")
	code.WriteString("\t\tfmt.Printf(\"AC: 所有测试通过 (%d/%d)\\n\", passed, total)\n")
	code.WriteString("\t} else {\n")
	code.WriteString("\t\tfmt.Printf(\"WA: 测试失败 (%d/%d)\\n\", passed, total)\n")
	code.WriteString("\t\tos.Exit(1)\n")
	code.WriteString("\t}\n")
	code.WriteString("}\n")

	return code.String()
}

// runCoreExecutable 运行核心模式可执行文件
func (s *JudgeService) runCoreExecutable(execFile, input string, timeLimit, memoryLimit int) *JudgeResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeLimit)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, execFile)

	// 设置输入
	cmd.Stdin = strings.NewReader(input)

	start := time.Now()
	output, err := cmd.CombinedOutput()
	executionTime := time.Since(start).Milliseconds()

	if ctx.Err() == context.DeadlineExceeded {
		return &JudgeResult{
			Status:        "TLE",
			ExecutionTime: executionTime,
			ErrorMsg:      "Time limit exceeded",
		}
	}

	// 对于核心模式，需要解析输出内容来判断是否通过
	outputStr := string(output)

	// 检查输出是否包含 "AC:" 表示所有测试通过
	if strings.Contains(outputStr, "AC:") {
		return &JudgeResult{
			Status:        "AC",
			ExecutionTime: executionTime,
			Output:        outputStr,
		}
	}

	// 检查输出是否包含 "WA:" 表示答案错误
	if strings.Contains(outputStr, "WA:") {
		return &JudgeResult{
			Status:        "WA",
			ExecutionTime: executionTime,
			Output:        outputStr,
			ErrorMsg:      "部分测试用例未通过",
		}
	}

	// 如果有错误退出码但没有明确的状态标识，检查是否是编译错误
	if err != nil {
		// 检查输出中是否包含测试结果
		if strings.Contains(outputStr, "PASS") || strings.Contains(outputStr, "FAIL") {
			// 包含测试输出，可能是WA
			return &JudgeResult{
				Status:        "WA",
				ExecutionTime: executionTime,
				Output:        outputStr,
				ErrorMsg:      "部分测试用例未通过",
			}
		} else {
			// 没有测试输出，可能是运行时错误
			return &JudgeResult{
				Status:        "RE",
				ExecutionTime: executionTime,
				ErrorMsg:      err.Error(),
				Output:        outputStr,
			}
		}
	}

	return &JudgeResult{
		Status:        "AC",
		ExecutionTime: executionTime,
		Output:        outputStr,
	}
}
