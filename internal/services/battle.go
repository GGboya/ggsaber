package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// BattleNotifier 接口用于发送对战通知
type BattleNotifier interface {
	NotifyBattleEnd(battleID string, room *BattleRoom)
	NotifyBattleFlee(battleID string, fleeUserID, opponentID uint)
}

type BattleService struct {
	redis       *redis.Client
	userService *UserService
	notifier    BattleNotifier // WebSocket 通知器
}

type BattleRoom struct {
	ID        string     `json:"id"`
	Player1ID uint       `json:"player1_id"`
	Player2ID uint       `json:"player2_id"`
	ProblemID uint       `json:"problem_id"`
	Status    string     `json:"status"` // waiting, active, finished
	StartTime time.Time  `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`

	Player1Ready bool `json:"player1_ready"`
	Player2Ready bool `json:"player2_ready"`

	Player1SubmitTime *time.Time `json:"player1_submit_time"`
	Player2SubmitTime *time.Time `json:"player2_submit_time"`
	Player1Code       string     `json:"player1_code"`
	Player2Code       string     `json:"player2_code"`
	Player1Result     string     `json:"player1_result"`
	Player2Result     string     `json:"player2_result"`

	WinnerID *uint `json:"winner_id"`
}

func NewBattleService(redis *redis.Client, userService *UserService) *BattleService {
	return &BattleService{
		redis:       redis,
		userService: userService,
	}
}

// SetNotifier 设置通知器
func (s *BattleService) SetNotifier(notifier BattleNotifier) {
	s.notifier = notifier
}

// GetBattleRoom 获取对战房间信息
func (s *BattleService) GetBattleRoom(battleID string) (*BattleRoom, error) {
	ctx := context.Background()

	data, err := s.redis.Get(ctx, fmt.Sprintf("battle:%s", battleID)).Result()
	if err != nil {
		return nil, err
	}

	// 首先尝试解析为完整的BattleRoom结构
	var room BattleRoom
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, err
	}

	// 如果是从匹配服务创建的简单数据，进行转换
	if room.ID == "" {
		var simpleData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &simpleData); err != nil {
			return nil, err
		}

		room = BattleRoom{
			ID:        battleID,
			Player1ID: uint(simpleData["player1"].(float64)),
			Player2ID: uint(simpleData["player2"].(float64)),
			Status:    simpleData["status"].(string),
		}

		// 如果有problem_id，设置它
		if problemID, ok := simpleData["problem_id"]; ok {
			room.ProblemID = uint(problemID.(float64))
		}
	}

	return &room, nil
}

// saveBattleRoom 保存对战房间信息
func (s *BattleService) saveBattleRoom(room *BattleRoom) error {
	ctx := context.Background()

	data, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, fmt.Sprintf("battle:%s", room.ID), string(data), 30*time.Minute).Err()
}

// JoinBattle 加入对战房间
func (s *BattleService) JoinBattle(battleID string, userID uint) error {
	room, err := s.GetBattleRoom(battleID)
	if err != nil {
		return err
	}

	// 检查用户是否属于这个房间
	if room.Player1ID != userID && room.Player2ID != userID {
		return fmt.Errorf("user not authorized for this battle")
	}

	// 设置用户准备状态
	if room.Player1ID == userID {
		room.Player1Ready = true
	} else {
		room.Player2Ready = true
	}

	// 如果双方都准备好了，开始对战
	if room.Player1Ready && room.Player2Ready && (room.Status == "waiting" || room.Status == "matched") {
		room.Status = "active"
		room.StartTime = time.Now()

		// 如果还没有题目，随机选择一个题目
		if room.ProblemID == 0 {
			problemID, err := s.selectRandomProblem()
			if err == nil {
				room.ProblemID = problemID
			}
		}

		// 通知双方对战开始
		s.notifyBattleStart(room)
	}

	return s.saveBattleRoom(room)
}

// SubmitCode 提交代码
func (s *BattleService) SubmitCode(battleID string, userID uint, code, language string) error {
	room, err := s.GetBattleRoom(battleID)
	if err != nil {
		return err
	}

	if room.Status != "active" {
		return fmt.Errorf("battle is not active")
	}

	now := time.Now()

	// 更新提交信息
	if room.Player1ID == userID {
		room.Player1Code = code
		room.Player1SubmitTime = &now
	} else if room.Player2ID == userID {
		room.Player2Code = code
		room.Player2SubmitTime = &now
	} else {
		return fmt.Errorf("user not authorized for this battle")
	}

	// 保存房间状态
	if err := s.saveBattleRoom(room); err != nil {
		return err
	}

	// 通知对手有人提交了代码
	s.notifyCodeSubmitted(room, userID)

	log.Printf("User %d submitted code for battle %s", userID, battleID)
	return nil
}

// UpdateResult 更新代码执行结果
func (s *BattleService) UpdateResult(battleID string, userID uint, result string, executionTime int64) error {
	room, err := s.GetBattleRoom(battleID)
	if err != nil {
		return err
	}

	// 更新结果
	if room.Player1ID == userID {
		room.Player1Result = result
	} else if room.Player2ID == userID {
		room.Player2Result = result
	} else {
		return fmt.Errorf("user not authorized for this battle")
	}

	// 检查是否可以结束对战
	if room.Player1Result != "" && room.Player2Result != "" {
		s.finishBattle(room)
	} else if result == "AC" {
		// 如果有人AC了，立即结束对战
		s.finishBattle(room)
	}

	return s.saveBattleRoom(room)
}

// finishBattle 结束对战
func (s *BattleService) finishBattle(room *BattleRoom) {
	now := time.Now()
	room.EndTime = &now
	room.Status = "finished"

	// 判断胜负
	winner := s.determineWinner(room)
	room.WinnerID = winner

	// 更新用户积分
	if winner != nil {
		s.updatePlayerRatings(room, *winner)
	}

	// 保存到数据库
	s.saveBattleToDatabase(room)

	// 通过 WebSocket 直接通知对战结束
	if s.notifier != nil {
		s.notifier.NotifyBattleEnd(room.ID, room)
	}

	// 同时发布 Redis 通知（兼容性）
	s.notifyBattleEnd(room)

	log.Printf("Battle %s finished, winner: %v", room.ID, winner)
}

// determineWinner 判断胜负
func (s *BattleService) determineWinner(room *BattleRoom) *uint {
	// 优先级：AC > 其他结果 > 未提交
	player1AC := room.Player1Result == "AC"
	player2AC := room.Player2Result == "AC"

	if player1AC && !player2AC {
		return &room.Player1ID
	}
	if player2AC && !player1AC {
		return &room.Player2ID
	}

	// 如果都AC了，比较提交时间
	if player1AC && player2AC {
		if room.Player1SubmitTime != nil && room.Player2SubmitTime != nil {
			if room.Player1SubmitTime.Before(*room.Player2SubmitTime) {
				return &room.Player1ID
			} else {
				return &room.Player2ID
			}
		}
	}

	// 其他情况暂时不分胜负
	return nil
}

// updatePlayerRatings 更新玩家积分
func (s *BattleService) updatePlayerRatings(room *BattleRoom, winnerID uint) {
	player1, err1 := s.userService.GetUserByID(room.Player1ID)
	player2, err2 := s.userService.GetUserByID(room.Player2ID)

	if err1 != nil || err2 != nil {
		log.Printf("Error getting users for rating update: %v, %v", err1, err2)
		return
	}

	var newRating1, newRating2 int
	if winnerID == room.Player1ID {
		newRating1, newRating2 = s.userService.CalculateNewRating(player1.Rating, player2.Rating)
		s.userService.UpdateRating(room.Player1ID, newRating1, true)
		s.userService.UpdateRating(room.Player2ID, newRating2, false)
	} else {
		newRating2, newRating1 = s.userService.CalculateNewRating(player2.Rating, player1.Rating)
		s.userService.UpdateRating(room.Player1ID, newRating1, false)
		s.userService.UpdateRating(room.Player2ID, newRating2, true)
	}

	log.Printf("Updated ratings: Player %d: %d->%d, Player %d: %d->%d",
		room.Player1ID, player1.Rating, newRating1,
		room.Player2ID, player2.Rating, newRating2)
}

// selectRandomProblem 随机选择题目
func (s *BattleService) selectRandomProblem() (uint, error) {
	// 这里简化处理，实际应该从数据库中选择
	// 可以根据用户水平选择合适难度的题目
	if time.Now().Unix()%2 == 0 {
		return 1, nil // 两数之和
	}
	return 2, nil // 反转链表
}

// saveBattleToDatabase 保存对战记录到数据库
func (s *BattleService) saveBattleToDatabase(room *BattleRoom) {
	// 这里应该保存到MySQL数据库
	// 由于简化实现，先跳过
	log.Printf("Should save battle %s to database", room.ID)
}

// notifyBattleStart 通知对战开始
func (s *BattleService) notifyBattleStart(room *BattleRoom) {
	ctx := context.Background()

	notification := map[string]interface{}{
		"type":       "battle_start",
		"battle_id":  room.ID,
		"problem_id": room.ProblemID,
		"start_time": room.StartTime.Unix(),
	}

	data, _ := json.Marshal(notification)
	s.redis.Publish(ctx, fmt.Sprintf("battle_notifications:%s", room.ID), string(data))
}

// notifyCodeSubmitted 通知代码提交
func (s *BattleService) notifyCodeSubmitted(room *BattleRoom, userID uint) {
	ctx := context.Background()

	notification := map[string]interface{}{
		"type":      "code_submitted",
		"battle_id": room.ID,
		"user_id":   userID,
	}

	data, _ := json.Marshal(notification)
	s.redis.Publish(ctx, fmt.Sprintf("battle_notifications:%s", room.ID), string(data))
}

// notifyBattleEnd 通知对战结束
func (s *BattleService) notifyBattleEnd(room *BattleRoom) {
	ctx := context.Background()

	notification := map[string]interface{}{
		"type":      "battle_end",
		"battle_id": room.ID,
		"winner_id": room.WinnerID,
		"end_time":  room.EndTime.Unix(),
	}

	data, _ := json.Marshal(notification)
	s.redis.Publish(ctx, fmt.Sprintf("battle_notifications:%s", room.ID), string(data))
}

// SetWebSocketHandler 设置 WebSocket 处理器（用于通知）
func (s *BattleService) SetWebSocketHandler(handler interface{}) {
	// 这里可以设置 WebSocket 处理器的引用，用于直接通知
	// 为了避免循环依赖，暂时保持简单实现
}

// FleeBattle 处理用户逃跑
func (s *BattleService) FleeBattle(battleID string, userID uint) error {
	room, err := s.GetBattleRoom(battleID)
	if err != nil {
		return err
	}

	// 检查用户是否属于这个房间
	if room.Player1ID != userID && room.Player2ID != userID {
		return fmt.Errorf("user not authorized for this battle")
	}

	// 只允许在对战进行中逃跑
	if room.Status != "active" {
		return fmt.Errorf("battle not active, cannot flee")
	}

	// 确定逃跑者和获胜者
	var fleeUserID, winnerID uint
	if room.Player1ID == userID {
		fleeUserID = room.Player1ID
		winnerID = room.Player2ID
	} else {
		fleeUserID = room.Player2ID
		winnerID = room.Player1ID
	}

	// 标记对战结束，设置获胜者
	now := time.Now()
	room.EndTime = &now
	room.Status = "finished"
	room.WinnerID = &winnerID

	// 标记逃跑者为逃跑状态
	if room.Player1ID == fleeUserID {
		room.Player1Result = "FLEE"
	} else {
		room.Player2Result = "FLEE"
	}

	// 标记获胜者为胜利状态
	if room.Player1ID == winnerID {
		room.Player1Result = "WIN_BY_FLEE"
	} else {
		room.Player2Result = "WIN_BY_FLEE"
	}

	// 更新玩家积分（逃跑者失败，对手获胜）
	s.updatePlayerRatings(room, winnerID)

	// 保存房间状态
	if err := s.saveBattleRoom(room); err != nil {
		return err
	}

	// 保存到数据库
	s.saveBattleToDatabase(room)

	// 通知双方逃跑事件
	if s.notifier != nil {
		s.notifier.NotifyBattleFlee(battleID, fleeUserID, winnerID)
	}

	// 清理匹配状态
	s.cleanupMatchStatus(room.Player1ID, room.Player2ID)

	log.Printf("User %d fled from battle %s, user %d wins", fleeUserID, battleID, winnerID)
	return nil
}

// cleanupMatchStatus 清理匹配状态
func (s *BattleService) cleanupMatchStatus(player1ID, player2ID uint) {
	ctx := context.Background()

	// 清理双方的匹配状态
	s.redis.Del(ctx, fmt.Sprintf("user_match:%d", player1ID))
	s.redis.Del(ctx, fmt.Sprintf("user_match:%d", player2ID))
}
