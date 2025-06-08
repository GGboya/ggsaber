package main

import (
	"go-saber-system/internal/api"
	"go-saber-system/internal/config"
	"go-saber-system/internal/database"
	"go-saber-system/internal/services"
	"log"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 初始化Redis
	redis := database.ConnectRedis(cfg.Redis)

	// 初始化服务
	userService := services.NewUserService(db, redis)
	matchService := services.NewMatchService(redis, userService)
	battleService := services.NewBattleService(redis, userService)
	problemService := services.NewProblemService(db)
	judgeService := services.NewJudgeService()

	// 启动匹配服务
	go matchService.Start()

	// 启动API服务器
	server := api.NewServer(cfg, userService, matchService, battleService, problemService, judgeService)

	log.Printf("Server starting on port %s", cfg.Server.Port)
	if err := server.Run(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
