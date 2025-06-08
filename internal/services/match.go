package services

import (
	"context"
	"encoding/json"
	"fmt"
	"go-saber-system/internal/models"
	"log"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type MatchService struct {
	redis       *redis.Client
	userService *UserService
	stopChan    chan bool
}

type MatchResult struct {
	BattleID string `json:"battle_id"`
	Player1  uint   `json:"player1"`
	Player2  uint   `json:"player2"`
}

func NewMatchService(redis *redis.Client, userService *UserService) *MatchService {
	return &MatchService{
		redis:       redis,
		userService: userService,
		stopChan:    make(chan bool),
	}
}

// Start 启动匹配服务
func (s *MatchService) Start() {
	ticker := time.NewTicker(2 * time.Second) // 每2秒检查一次匹配
	defer ticker.Stop()

	log.Println("Match service started")

	for {
		select {
		case <-ticker.C:
			s.processMatching()
		case <-s.stopChan:
			log.Println("Match service stopped")
			return
		}
	}
}

// Stop 停止匹配服务
func (s *MatchService) Stop() {
	s.stopChan <- true
}

// JoinQueue 加入匹配队列
func (s *MatchService) JoinQueue(userID uint) error {
	ctx := context.Background()

	// 检查用户是否已经在队列中
	inQueue, err := s.userService.IsUserInMatch(userID)
	if err != nil {
		return err
	}
	if inQueue {
		return fmt.Errorf("user already in queue")
	}

	// 获取用户信息
	user, err := s.userService.GetUserByID(userID)
	if err != nil {
		return err
	}

	// 创建匹配请求
	matchReq := models.MatchRequest{
		UserID:    userID,
		Rating:    user.Rating,
		QueueTime: time.Now(),
	}

	// 序列化并存储到Redis
	data, err := json.Marshal(matchReq)
	if err != nil {
		return err
	}

	// 添加到匹配队列
	err = s.redis.ZAdd(ctx, "match_queue", redis.Z{
		Score:  float64(user.Rating),
		Member: string(data),
	}).Err()
	if err != nil {
		return err
	}

	// 设置用户匹配状态
	err = s.redis.Set(ctx, fmt.Sprintf("match_queue:%d", userID), "waiting", 10*time.Minute).Err()
	if err != nil {
		return err
	}

	log.Printf("User %d joined match queue with rating %d", userID, user.Rating)
	return nil
}

// LeaveQueue 离开匹配队列
func (s *MatchService) LeaveQueue(userID uint) error {
	ctx := context.Background()

	// 从有序集合中移除用户
	members, err := s.redis.ZRange(ctx, "match_queue", 0, -1).Result()
	if err != nil {
		return err
	}

	for _, member := range members {
		var matchReq models.MatchRequest
		if err := json.Unmarshal([]byte(member), &matchReq); err != nil {
			continue
		}

		if matchReq.UserID == userID {
			s.redis.ZRem(ctx, "match_queue", member)
			break
		}
	}

	// 删除用户匹配状态
	s.redis.Del(ctx, fmt.Sprintf("match_queue:%d", userID))

	log.Printf("User %d left match queue", userID)
	return nil
}

// processMatching 处理匹配逻辑
func (s *MatchService) processMatching() {
	ctx := context.Background()

	// 获取所有等待匹配的用户
	members, err := s.redis.ZRange(ctx, "match_queue", 0, -1).Result()
	if err != nil {
		log.Printf("Error getting match queue: %v", err)
		return
	}

	if len(members) < 2 {
		return // 不足两人，无法匹配
	}

	// 解析用户数据
	var requests []models.MatchRequest
	for _, member := range members {
		var req models.MatchRequest
		if err := json.Unmarshal([]byte(member), &req); err != nil {
			continue
		}
		requests = append(requests, req)
	}

	// 查找匹配
	matches := s.findMatches(requests)

	// 处理匹配结果
	for _, match := range matches {
		s.createBattle(match)
	}
}

// findMatches 查找合适的匹配
func (s *MatchService) findMatches(requests []models.MatchRequest) []MatchResult {
	var matches []MatchResult
	used := make(map[uint]bool)

	for i := 0; i < len(requests); i++ {
		if used[requests[i].UserID] {
			continue
		}

		for j := i + 1; j < len(requests); j++ {
			if used[requests[j].UserID] {
				continue
			}

			// 检查积分差距和等待时间
			if s.isGoodMatch(requests[i], requests[j]) {
				battleID := uuid.New().String()
				matches = append(matches, MatchResult{
					BattleID: battleID,
					Player1:  requests[i].UserID,
					Player2:  requests[j].UserID,
				})

				used[requests[i].UserID] = true
				used[requests[j].UserID] = true
				break
			}
		}
	}

	return matches
}

// isGoodMatch 判断两个用户是否适合匹配
func (s *MatchService) isGoodMatch(req1, req2 models.MatchRequest) bool {
	// 计算积分差距
	ratingDiff := math.Abs(float64(req1.Rating - req2.Rating))

	// 计算等待时间
	waitTime1 := time.Since(req1.QueueTime).Seconds()
	waitTime2 := time.Since(req2.QueueTime).Seconds()
	maxWaitTime := math.Max(waitTime1, waitTime2)

	// 基础积分差距限制
	baseRatingDiff := 100.0

	// 等待时间越长，积分差距限制越宽松
	adjustedRatingDiff := baseRatingDiff + (maxWaitTime / 10.0 * 50.0)

	return ratingDiff <= adjustedRatingDiff
}

// createBattle 创建对战房间
func (s *MatchService) createBattle(match MatchResult) {
	ctx := context.Background()

	// 从队列中移除已匹配的用户
	s.LeaveQueue(match.Player1)
	s.LeaveQueue(match.Player2)

	// 随机选择一个题目给这场对战
	problemID := s.selectProblemForBattle()

	// 创建对战房间数据
	battleData := map[string]interface{}{
		"battle_id":  match.BattleID,
		"player1":    match.Player1,
		"player2":    match.Player2,
		"problem_id": problemID,
		"status":     "matched",
		"created_at": time.Now().Unix(),
	}

	// 存储对战房间信息
	data, _ := json.Marshal(battleData)
	s.redis.Set(ctx, fmt.Sprintf("battle:%s", match.BattleID), string(data), 30*time.Minute)

	// 通知用户匹配成功
	s.notifyMatchSuccess(match)

	log.Printf("Created battle %s between users %d and %d", match.BattleID, match.Player1, match.Player2)
}

// notifyMatchSuccess 通知用户匹配成功
func (s *MatchService) notifyMatchSuccess(match MatchResult) {
	ctx := context.Background()

	// 设置用户状态为已匹配
	s.redis.Set(ctx, fmt.Sprintf("user_match:%d", match.Player1), match.BattleID, 30*time.Minute)
	s.redis.Set(ctx, fmt.Sprintf("user_match:%d", match.Player2), match.BattleID, 30*time.Minute)

	// 发布匹配成功消息（用于WebSocket通知）
	notification := map[string]interface{}{
		"type":      "match_success",
		"battle_id": match.BattleID,
	}

	data, _ := json.Marshal(notification)
	s.redis.Publish(ctx, fmt.Sprintf("user_notifications:%d", match.Player1), string(data))
	s.redis.Publish(ctx, fmt.Sprintf("user_notifications:%d", match.Player2), string(data))
}

// selectProblemForBattle 为对战选择题目
func (s *MatchService) selectProblemForBattle() uint {
	// 固定使用反转链表题目（核心模式）
	return 3 // 反转链表 (ID=3，由add_reverse_linked_list.go脚本创建)
}

// GetMatchStatus 获取用户匹配状态
func (s *MatchService) GetMatchStatus(userID uint) (string, error) {
	ctx := context.Background()

	// 检查是否在队列中
	inQueue, err := s.userService.IsUserInMatch(userID)
	if err != nil {
		return "", err
	}
	if inQueue {
		return "waiting", nil
	}

	// 检查是否已匹配
	battleID, err := s.redis.Get(ctx, fmt.Sprintf("user_match:%d", userID)).Result()
	if err == redis.Nil {
		return "idle", nil
	}
	if err != nil {
		return "", err
	}

	return battleID, nil
}
