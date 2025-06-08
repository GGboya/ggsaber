package api

import (
	"net/http"

	"go-saber-system/internal/config"
	"go-saber-system/internal/services"

	"github.com/gin-gonic/gin"
)

type Server struct {
	config         *config.Config
	router         *gin.Engine
	userService    *services.UserService
	matchService   *services.MatchService
	battleService  *services.BattleService
	problemService *services.ProblemService
	judgeService   *services.JudgeService
	wsHandler      *WebSocketHandler
}

func NewServer(
	cfg *config.Config,
	userService *services.UserService,
	matchService *services.MatchService,
	battleService *services.BattleService,
	problemService *services.ProblemService,
	judgeService *services.JudgeService,
) *Server {
	server := &Server{
		config:         cfg,
		userService:    userService,
		matchService:   matchService,
		battleService:  battleService,
		problemService: problemService,
		judgeService:   judgeService,
	}

	server.wsHandler = NewWebSocketHandler(userService, matchService, battleService)

	// 设置 WebSocket 通知器，让 battleService 能够直接通知客户端
	battleService.SetNotifier(server.wsHandler)

	server.setupRoutes()

	return server
}

func (s *Server) setupRoutes() {
	s.router = gin.Default()

	// CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务
	s.router.Static("/static", "./web")
	s.router.StaticFile("/", "./web/index.html")
	s.router.StaticFile("/favicon.ico", "./web/favicon.ico")

	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// WebSocket endpoint
	s.router.GET("/ws", s.wsHandler.HandleWebSocket)

	// API routes
	api := s.router.Group("/api/v1")
	{
		// User routes
		users := api.Group("/users")
		{
			users.POST("/register", s.handleRegister)
			users.POST("/login", s.handleLogin)
			users.GET("/profile/:id", s.handleGetProfile)
			users.GET("/leaderboard", s.handleGetLeaderboard)
		}

		// Match routes
		match := api.Group("/match")
		{
			match.POST("/join", s.handleJoinMatch)
			match.DELETE("/leave", s.handleLeaveMatch)
			match.GET("/status", s.handleGetMatchStatus)
		}

		// Battle routes
		battle := api.Group("/battle")
		{
			battle.GET("/:id", s.handleGetBattle)
			battle.POST("/:id/join", s.handleJoinBattle)
			battle.POST("/:id/submit", s.handleSubmitCode)
			battle.POST("/:id/flee", s.handleFleeBattle)
		}

		// Problem routes
		problems := api.Group("/problems")
		{
			problems.GET("", s.handleListProblems)
			problems.GET("/:id", s.handleGetProblem)
			problems.POST("", s.handleCreateProblem)
		}

		// Judge routes
		judge := api.Group("/judge")
		{
			judge.POST("/check", s.handleCheckSyntax)
			judge.GET("/languages", s.handleGetLanguages)
		}
	}
}

func (s *Server) Run() error {
	return s.router.Run(":" + s.config.Server.Port)
}
