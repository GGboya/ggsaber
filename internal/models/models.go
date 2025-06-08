package models

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;not null;size:50" json:"username"`
	Email    string `gorm:"uniqueIndex;not null;size:100" json:"email"`
	Password string `gorm:"not null;size:64" json:"-"`
	Rating   int    `gorm:"default:1000" json:"rating"` // ELO积分
	Wins     int    `gorm:"default:0" json:"wins"`
	Losses   int    `gorm:"default:0" json:"losses"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Problem 题目模型
type Problem struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Title       string `gorm:"not null;size:255" json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Difficulty  string `gorm:"not null;size:20" json:"difficulty"` // easy, medium, hard
	TimeLimit   int    `gorm:"default:1000" json:"time_limit"`     // 毫秒
	MemoryLimit int    `gorm:"default:256" json:"memory_limit"`    // MB

	// 题目模式：traditional(传统main函数) 或 core(力扣核心模式)
	Mode string `gorm:"default:traditional;size:20" json:"mode"`

	// JSON存储测试用例
	TestCases string `gorm:"type:text" json:"test_cases"`

	// 核心模式的代码模板(JSON格式存储各语言模板)
	CodeTemplates string `gorm:"type:text" json:"code_templates"`

	// 核心模式的函数签名信息
	FunctionSignature string `gorm:"type:text" json:"function_signature"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Battle 对战记录
type Battle struct {
	ID        string `gorm:"primaryKey;size:36" json:"id"`
	Player1ID uint   `gorm:"not null" json:"player1_id"`
	Player2ID uint   `gorm:"not null" json:"player2_id"`
	ProblemID uint   `gorm:"not null" json:"problem_id"`

	Player1 User    `gorm:"foreignKey:Player1ID" json:"player1"`
	Player2 User    `gorm:"foreignKey:Player2ID" json:"player2"`
	Problem Problem `gorm:"foreignKey:ProblemID" json:"problem"`

	Status    string     `gorm:"default:waiting;size:20" json:"status"` // waiting, active, finished
	WinnerID  *uint      `json:"winner_id"`
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	// 对战详细数据
	Player1SubmitTime *time.Time `json:"player1_submit_time"`
	Player2SubmitTime *time.Time `json:"player2_submit_time"`
	Player1Code       string     `gorm:"type:text" json:"player1_code"`
	Player2Code       string     `gorm:"type:text" json:"player2_code"`
	Player1Result     string     `gorm:"size:10" json:"player1_result"` // AC, WA, TLE, MLE, CE, RE
	Player2Result     string     `gorm:"size:10" json:"player2_result"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Submission 代码提交记录
type Submission struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"not null" json:"user_id"`
	BattleID string `gorm:"not null;size:36" json:"battle_id"`
	Code     string `gorm:"type:text" json:"code"`
	Language string `gorm:"not null;size:20" json:"language"`
	Result   string `gorm:"size:10" json:"result"` // AC(Accepted), WA(Wrong Answer), TLE(Time Limit Exceeded), etc.

	ExecutionTime int64  `json:"execution_time"` // 执行时间(毫秒)
	MemoryUsage   int64  `json:"memory_usage"`   // 内存使用(KB)
	ErrorMsg      string `gorm:"type:text" json:"error_msg"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TestCase 测试用例结构（用于JSON序列化）
type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// CoreTestCase 核心模式测试用例结构
type CoreTestCase struct {
	Args     []interface{} `json:"args"`     // 函数参数
	Expected interface{}   `json:"expected"` // 期望输出
	Explain  string        `json:"explain"`  // 用例说明
}

// CodeTemplate 代码模板结构
type CodeTemplate struct {
	Go     string `json:"go"`
	Python string `json:"python"`
	CPP    string `json:"cpp"`
	Java   string `json:"java"`
}

// FunctionSignature 函数签名信息
type FunctionSignature struct {
	FunctionName string            `json:"function_name"`
	ReturnType   map[string]string `json:"return_type"` // 各语言的返回类型
	Parameters   []Parameter       `json:"parameters"`
}

// Parameter 函数参数定义
type Parameter struct {
	Name     string            `json:"name"`
	Type     map[string]string `json:"type"`     // 各语言的参数类型
	Describe string            `json:"describe"` // 参数说明
}

// MatchRequest 匹配请求
type MatchRequest struct {
	UserID    uint      `json:"user_id"`
	Rating    int       `json:"rating"`
	QueueTime time.Time `json:"queue_time"`
}
