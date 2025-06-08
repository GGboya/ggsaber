package api

import (
	"net/http"
	"strconv"

	"go-saber-system/internal/models"
	"go-saber-system/internal/services"

	"github.com/gin-gonic/gin"
)

// User handlers
func (s *Server) handleRegister(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": user})
}

func (s *Server) handleLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := s.userService.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *Server) handleGetProfile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := s.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func (s *Server) handleGetLeaderboard(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	users, err := s.userService.GetLeaderboard(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"leaderboard": users})
}

// Match handlers
func (s *Server) handleJoinMatch(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.matchService.JoinQueue(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined match queue successfully"})
}

func (s *Server) handleLeaveMatch(c *gin.Context) {
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.matchService.LeaveQueue(req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Left match queue successfully"})
}

func (s *Server) handleGetMatchStatus(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	status, err := s.matchService.GetMatchStatus(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": status})
}

// Battle handlers
func (s *Server) handleGetBattle(c *gin.Context) {
	battleID := c.Param("id")

	room, err := s.battleService.GetBattleRoom(battleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Battle not found"})
		return
	}

	// 获取两个玩家的详细信息
	player1, err := s.userService.GetUserByID(room.Player1ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get player1 info"})
		return
	}

	player2, err := s.userService.GetUserByID(room.Player2ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get player2 info"})
		return
	}

	// 构建包含完整玩家信息的响应
	battleData := map[string]interface{}{
		"id":         room.ID,
		"problem_id": room.ProblemID,
		"status":     room.Status,
		"start_time": room.StartTime,
		"end_time":   room.EndTime,
		"player1": map[string]interface{}{
			"id":          player1.ID,
			"username":    player1.Username,
			"rating":      player1.Rating,
			"ready":       room.Player1Ready,
			"code":        room.Player1Code,
			"result":      room.Player1Result,
			"submit_time": room.Player1SubmitTime,
		},
		"player2": map[string]interface{}{
			"id":          player2.ID,
			"username":    player2.Username,
			"rating":      player2.Rating,
			"ready":       room.Player2Ready,
			"code":        room.Player2Code,
			"result":      room.Player2Result,
			"submit_time": room.Player2SubmitTime,
		},
		"winner_id": room.WinnerID,
	}

	c.JSON(http.StatusOK, gin.H{"battle": battleData})
}

func (s *Server) handleJoinBattle(c *gin.Context) {
	battleID := c.Param("id")

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.battleService.JoinBattle(battleID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Joined battle successfully"})
}

func (s *Server) handleSubmitCode(c *gin.Context) {
	battleID := c.Param("id")

	var req struct {
		UserID   uint   `json:"user_id" binding:"required"`
		Code     string `json:"code" binding:"required"`
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 提交代码到对战房间
	if err := s.battleService.SubmitCode(battleID, req.UserID, req.Code, req.Language); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 异步判定代码
	go s.judgeCodeAsync(battleID, req.UserID, req.Code, req.Language)

	c.JSON(http.StatusOK, gin.H{"message": "Code submitted successfully"})
}

func (s *Server) handleFleeBattle(c *gin.Context) {
	battleID := c.Param("id")

	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 处理逃跑
	if err := s.battleService.FleeBattle(battleID, req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Battle fled successfully"})
}

// Problem handlers
func (s *Server) handleListProblems(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")
	difficulty := c.Query("difficulty")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	problems, total, err := s.problemService.ListProblems(page, pageSize, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"problems":  problems,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (s *Server) handleGetProblem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	problem, err := s.problemService.GetProblemByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
		return
	}

	// 构建响应数据
	response := gin.H{"problem": problem}

	// 如果是核心模式，添加额外的数据
	if problem.Mode == "core" {
		templates, err := s.problemService.GetCodeTemplates(uint(id))
		if err == nil {
			response["code_templates"] = templates
		}

		signature, err := s.problemService.GetFunctionSignature(uint(id))
		if err == nil {
			response["function_signature"] = signature
		}

		coreTestCases, err := s.problemService.GetCoreTestCases(uint(id))
		if err == nil {
			response["core_test_cases"] = coreTestCases
		}
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleCreateProblem(c *gin.Context) {
	var req struct {
		Title       string            `json:"title" binding:"required"`
		Description string            `json:"description" binding:"required"`
		Difficulty  string            `json:"difficulty" binding:"required"`
		TimeLimit   int               `json:"time_limit" binding:"required"`
		MemoryLimit int               `json:"memory_limit" binding:"required"`
		TestCases   []models.TestCase `json:"test_cases" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	problem, err := s.problemService.CreateProblem(
		req.Title,
		req.Description,
		req.Difficulty,
		req.TimeLimit,
		req.MemoryLimit,
		req.TestCases,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"problem": problem})
}

// Judge handlers
func (s *Server) handleCheckSyntax(c *gin.Context) {
	var req struct {
		Code     string `json:"code" binding:"required"`
		Language string `json:"language" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.judgeService.CheckSyntax(req.Code, req.Language); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Syntax is valid"})
}

func (s *Server) handleGetLanguages(c *gin.Context) {
	languages := s.judgeService.GetSupportedLanguages()
	c.JSON(http.StatusOK, gin.H{"languages": languages})
}

// judgeCodeAsync 异步判定代码
func (s *Server) judgeCodeAsync(battleID string, userID uint, code, language string) {
	// 获取对战房间信息
	room, err := s.battleService.GetBattleRoom(battleID)
	if err != nil {
		return
	}

	// 获取题目信息
	problem, err := s.problemService.GetProblemByID(room.ProblemID)
	if err != nil {
		return
	}

	var result *services.JudgeResult

	// 根据题目模式选择判题方式
	if problem.Mode == "core" {
		// 核心模式：使用函数调用测试
		coreTestCases, err := s.problemService.GetCoreTestCases(problem.ID)
		if err != nil {
			s.notifyJudgeError(userID, "获取测试用例失败", err.Error())
			return
		}

		signature, err := s.problemService.GetFunctionSignature(problem.ID)
		if err != nil {
			s.notifyJudgeError(userID, "获取函数签名失败", err.Error())
			return
		}

		// 执行核心模式判定
		result, err = s.judgeService.JudgeCoreCode(code, language, coreTestCases, signature, problem.TimeLimit, problem.MemoryLimit)
	} else {
		// 传统模式：使用输入输出测试
		testCases, err := s.problemService.GetTestCases(problem.ID)
		if err != nil {
			s.notifyJudgeError(userID, "获取测试用例失败", err.Error())
			return
		}

		// 执行传统模式判定
		result, err = s.judgeService.JudgeCode(code, language, testCases, problem.TimeLimit, problem.MemoryLimit)
	}

	if err != nil {
		s.notifyJudgeError(userID, "判题系统错误", err.Error())
		return
	}

	// 更新对战结果
	s.battleService.UpdateResult(battleID, userID, result.Status, result.ExecutionTime)

	// 通过WebSocket通知判题结果（只发给提交代码的用户）
	if s.wsHandler != nil {
		s.wsHandler.SendToUser(userID, WSMessage{
			Type: "judge_result",
			Data: map[string]interface{}{
				"user_id":        userID,
				"status":         result.Status,
				"execution_time": result.ExecutionTime,
				"memory_usage":   result.MemoryUsage,
				"message":        getStatusMessage(result.Status),
				"output":         result.Output,
				"error_msg":      result.ErrorMsg,
			},
		})
	}
}

// notifyJudgeError 通知判题错误
func (s *Server) notifyJudgeError(userID uint, message, error string) {
	if s.wsHandler != nil {
		s.wsHandler.SendToUser(userID, WSMessage{
			Type: "judge_result",
			Data: map[string]interface{}{
				"user_id": userID,
				"status":  "Error",
				"message": message,
				"error":   error,
			},
		})
	}
}

// getStatusMessage 获取状态消息
func getStatusMessage(status string) string {
	switch status {
	case "Accepted":
		return "恭喜！代码通过所有测试用例"
	case "Wrong Answer":
		return "答案错误，请检查逻辑"
	case "Time Limit Exceeded":
		return "超出时间限制"
	case "Memory Limit Exceeded":
		return "超出内存限制"
	case "Runtime Error":
		return "运行时错误"
	case "Compile Error":
		return "编译错误"
	default:
		return "判题完成"
	}
}
