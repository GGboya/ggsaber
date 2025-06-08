package services

import (
	"encoding/json"
	"go-saber-system/internal/models"
	"math/rand"

	"gorm.io/gorm"
)

type ProblemService struct {
	db *gorm.DB
}

func NewProblemService(db *gorm.DB) *ProblemService {
	return &ProblemService{
		db: db,
	}
}

// GetProblemByID 根据ID获取题目
func (s *ProblemService) GetProblemByID(id uint) (*models.Problem, error) {
	var problem models.Problem
	err := s.db.First(&problem, id).Error
	if err != nil {
		return nil, err
	}
	return &problem, nil
}

// GetRandomProblem 随机获取题目
func (s *ProblemService) GetRandomProblem(difficulty string) (*models.Problem, error) {
	var problems []models.Problem
	query := s.db

	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	err := query.Find(&problems).Error
	if err != nil {
		return nil, err
	}

	if len(problems) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// 随机选择一个题目
	randomIndex := rand.Intn(len(problems))
	return &problems[randomIndex], nil
}

// GetTestCases 获取题目的测试用例
func (s *ProblemService) GetTestCases(problemID uint) ([]models.TestCase, error) {
	problem, err := s.GetProblemByID(problemID)
	if err != nil {
		return nil, err
	}

	var testCases []models.TestCase
	if err := json.Unmarshal([]byte(problem.TestCases), &testCases); err != nil {
		return nil, err
	}

	return testCases, nil
}

// CreateProblem 创建新题目
func (s *ProblemService) CreateProblem(title, description, difficulty string, timeLimit, memoryLimit int, testCases []models.TestCase) (*models.Problem, error) {
	testCasesJSON, err := json.Marshal(testCases)
	if err != nil {
		return nil, err
	}

	problem := &models.Problem{
		Title:       title,
		Description: description,
		Difficulty:  difficulty,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit,
		TestCases:   string(testCasesJSON),
		Mode:        "traditional",
	}

	if err := s.db.Create(problem).Error; err != nil {
		return nil, err
	}

	return problem, nil
}

// CreateCoreProblem 创建核心模式题目
func (s *ProblemService) CreateCoreProblem(title, description, difficulty string, timeLimit, memoryLimit int,
	coreTestCases []models.CoreTestCase, codeTemplates models.CodeTemplate, functionSignature models.FunctionSignature) (*models.Problem, error) {

	testCasesJSON, err := json.Marshal(coreTestCases)
	if err != nil {
		return nil, err
	}

	templatesJSON, err := json.Marshal(codeTemplates)
	if err != nil {
		return nil, err
	}

	signatureJSON, err := json.Marshal(functionSignature)
	if err != nil {
		return nil, err
	}

	problem := &models.Problem{
		Title:             title,
		Description:       description,
		Difficulty:        difficulty,
		TimeLimit:         timeLimit,
		MemoryLimit:       memoryLimit,
		Mode:              "core",
		TestCases:         string(testCasesJSON),
		CodeTemplates:     string(templatesJSON),
		FunctionSignature: string(signatureJSON),
	}

	if err := s.db.Create(problem).Error; err != nil {
		return nil, err
	}

	return problem, nil
}

// GetCoreTestCases 获取核心模式测试用例
func (s *ProblemService) GetCoreTestCases(problemID uint) ([]models.CoreTestCase, error) {
	problem, err := s.GetProblemByID(problemID)
	if err != nil {
		return nil, err
	}

	if problem.Mode != "core" {
		return nil, gorm.ErrRecordNotFound
	}

	var testCases []models.CoreTestCase
	if err := json.Unmarshal([]byte(problem.TestCases), &testCases); err != nil {
		return nil, err
	}

	return testCases, nil
}

// GetCodeTemplates 获取代码模板
func (s *ProblemService) GetCodeTemplates(problemID uint) (models.CodeTemplate, error) {
	problem, err := s.GetProblemByID(problemID)
	if err != nil {
		return models.CodeTemplate{}, err
	}

	if problem.Mode != "core" {
		return models.CodeTemplate{}, gorm.ErrRecordNotFound
	}

	var templates models.CodeTemplate
	if err := json.Unmarshal([]byte(problem.CodeTemplates), &templates); err != nil {
		return models.CodeTemplate{}, err
	}

	return templates, nil
}

// GetFunctionSignature 获取函数签名
func (s *ProblemService) GetFunctionSignature(problemID uint) (models.FunctionSignature, error) {
	problem, err := s.GetProblemByID(problemID)
	if err != nil {
		return models.FunctionSignature{}, err
	}

	if problem.Mode != "core" {
		return models.FunctionSignature{}, gorm.ErrRecordNotFound
	}

	var signature models.FunctionSignature
	if err := json.Unmarshal([]byte(problem.FunctionSignature), &signature); err != nil {
		return models.FunctionSignature{}, err
	}

	return signature, nil
}

// ListProblems 获取题目列表
func (s *ProblemService) ListProblems(page, pageSize int, difficulty string) ([]models.Problem, int64, error) {
	var problems []models.Problem
	var total int64

	query := s.db.Model(&models.Problem{})
	if difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, total, nil
}
