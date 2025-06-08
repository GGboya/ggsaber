package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go-saber-system/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketClient 包装 WebSocket 连接和相关信息
type WebSocketClient struct {
	conn   *websocket.Conn
	mutex  sync.Mutex // 保护写入操作的互斥锁
	userID uint
}

type WebSocketHandler struct {
	upgrader      websocket.Upgrader
	clients       map[uint]*WebSocketClient
	clientsMutex  sync.RWMutex
	userService   *services.UserService
	matchService  *services.MatchService
	battleService *services.BattleService
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func NewWebSocketHandler(
	userService *services.UserService,
	matchService *services.MatchService,
	battleService *services.BattleService,
) *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该更严格
			},
		},
		clients:       make(map[uint]*WebSocketClient),
		userService:   userService,
		matchService:  matchService,
		battleService: battleService,
	}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// 从查询参数中获取用户ID
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userIDUint, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	userID := uint(userIDUint)

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	log.Printf("User %d connected via WebSocket", userID)

	// 注册客户端
	client := &WebSocketClient{
		conn:   conn,
		userID: userID,
	}
	h.registerClient(userID, client)

	// 启动通知监听器
	go h.listenForNotifications(userID)

	// 处理WebSocket消息
	defer func() {
		h.unregisterClient(userID)
		conn.Close()
		log.Printf("User %d disconnected from WebSocket", userID)
	}()

	for {
		var msg WSMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if !websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for user %d: %v", userID, err)
			}
			break
		}

		h.handleMessage(userID, msg)
	}
}

func (h *WebSocketHandler) registerClient(userID uint, client *WebSocketClient) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	h.clients[userID] = client
}

func (h *WebSocketHandler) unregisterClient(userID uint) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	delete(h.clients, userID)
}

func (h *WebSocketHandler) sendToClient(userID uint, message WSMessage) {
	h.clientsMutex.RLock()
	client, exists := h.clients[userID]
	h.clientsMutex.RUnlock()

	if !exists {
		return
	}

	// 使用客户端的互斥锁保护写入操作
	client.mutex.Lock()
	defer client.mutex.Unlock()

	if err := client.conn.WriteJSON(message); err != nil {
		log.Printf("Failed to send message to user %d: %v", userID, err)
		h.unregisterClient(userID)
	}
}

func (h *WebSocketHandler) handleMessage(userID uint, msg WSMessage) {
	switch msg.Type {
	case "join_match":
		h.handleJoinMatch(userID)
	case "leave_match":
		h.handleLeaveMatch(userID)
	case "ping":
		h.sendToClient(userID, WSMessage{Type: "pong", Data: nil})
	case "battle_ready":
		h.handleBattleReady(userID, msg.Data)
	case "battle_code":
		h.handleBattleCode(userID, msg.Data)
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

func (h *WebSocketHandler) handleJoinMatch(userID uint) {
	if err := h.matchService.JoinQueue(userID); err != nil {
		h.sendToClient(userID, WSMessage{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	h.sendToClient(userID, WSMessage{
		Type: "match_joined",
		Data: map[string]string{"status": "waiting"},
	})
}

func (h *WebSocketHandler) handleLeaveMatch(userID uint) {
	if err := h.matchService.LeaveQueue(userID); err != nil {
		h.sendToClient(userID, WSMessage{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
		return
	}

	h.sendToClient(userID, WSMessage{
		Type: "match_left",
		Data: map[string]string{"status": "idle"},
	})
}

func (h *WebSocketHandler) handleBattleReady(userID uint, data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	battleID, ok := dataMap["battle_id"].(string)
	if !ok {
		return
	}

	if err := h.battleService.JoinBattle(battleID, userID); err != nil {
		h.sendToClient(userID, WSMessage{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
	}
}

func (h *WebSocketHandler) handleBattleCode(userID uint, data interface{}) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return
	}

	battleID, _ := dataMap["battle_id"].(string)
	code, _ := dataMap["code"].(string)
	language, _ := dataMap["language"].(string)

	if battleID == "" || code == "" || language == "" {
		return
	}

	if err := h.battleService.SubmitCode(battleID, userID, code, language); err != nil {
		h.sendToClient(userID, WSMessage{
			Type: "error",
			Data: map[string]string{"message": err.Error()},
		})
	}
}

func (h *WebSocketHandler) listenForNotifications(userID uint) {
	ticker := time.NewTicker(3 * time.Second) // 减少检查频率
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查客户端是否还连接
			h.clientsMutex.RLock()
			_, exists := h.clients[userID]
			h.clientsMutex.RUnlock()

			if !exists {
				return // 客户端已断开连接，退出goroutine
			}

			h.checkUserStatus(userID)
		}
	}
}

func (h *WebSocketHandler) checkUserStatus(userID uint) {
	// 检查匹配状态
	status, err := h.matchService.GetMatchStatus(userID)
	if err != nil {
		return
	}

	// 如果状态改变，通知客户端
	if status != "idle" && status != "waiting" {
		// 匹配成功，status是battleID
		h.sendToClient(userID, WSMessage{
			Type: "match_success",
			Data: map[string]string{"battle_id": status},
		})
	}
}

func (h *WebSocketHandler) BroadcastToBattle(battleID string, message WSMessage) {
	room, err := h.battleService.GetBattleRoom(battleID)
	if err != nil {
		return
	}

	h.sendToClient(room.Player1ID, message)
	h.sendToClient(room.Player2ID, message)
}

// SendToUser 发送消息给特定用户
func (h *WebSocketHandler) SendToUser(userID uint, message WSMessage) {
	h.sendToClient(userID, message)
}

func (h *WebSocketHandler) NotifyMatchSuccess(userID1, userID2 uint, battleID string) {
	message := WSMessage{
		Type: "match_success",
		Data: map[string]string{"battle_id": battleID},
	}

	h.sendToClient(userID1, message)
	h.sendToClient(userID2, message)
}

func (h *WebSocketHandler) NotifyBattleStart(battleID string, problemID uint) {
	message := WSMessage{
		Type: "battle_start",
		Data: map[string]interface{}{
			"battle_id":  battleID,
			"problem_id": problemID,
		},
	}

	h.BroadcastToBattle(battleID, message)
}

func (h *WebSocketHandler) NotifyBattleEnd(battleID string, room *services.BattleRoom) {
	// 获取玩家信息用于显示结果
	player1, _ := h.userService.GetUserByID(room.Player1ID)
	player2, _ := h.userService.GetUserByID(room.Player2ID)

	var winner *map[string]interface{}
	var loser *map[string]interface{}

	if room.WinnerID != nil {
		if *room.WinnerID == room.Player1ID {
			winner = &map[string]interface{}{
				"id":       room.Player1ID,
				"username": player1.Username,
				"rating":   player1.Rating,
			}
			loser = &map[string]interface{}{
				"id":       room.Player2ID,
				"username": player2.Username,
				"rating":   player2.Rating,
			}
		} else {
			winner = &map[string]interface{}{
				"id":       room.Player2ID,
				"username": player2.Username,
				"rating":   player2.Rating,
			}
			loser = &map[string]interface{}{
				"id":       room.Player1ID,
				"username": player1.Username,
				"rating":   player1.Rating,
			}
		}
	}

	// 计算对战时长
	var battleDuration int64
	if room.EndTime != nil {
		battleDuration = int64(room.EndTime.Sub(room.StartTime).Seconds())
	}

	message := WSMessage{
		Type: "battle_end",
		Data: map[string]interface{}{
			"battle_id":           battleID,
			"winner":              winner,
			"loser":               loser,
			"battle_duration":     battleDuration,
			"player1_result":      room.Player1Result,
			"player2_result":      room.Player2Result,
			"player1_submit_time": room.Player1SubmitTime,
			"player2_submit_time": room.Player2SubmitTime,
		},
	}

	h.BroadcastToBattle(battleID, message)
	log.Printf("Notified battle end for %s, winner: %v", battleID, room.WinnerID)
}

func (h *WebSocketHandler) NotifyCodeSubmitted(battleID string, userID uint) {
	message := WSMessage{
		Type: "code_submitted",
		Data: map[string]interface{}{
			"battle_id": battleID,
			"user_id":   userID,
		},
	}

	h.BroadcastToBattle(battleID, message)
}

func (h *WebSocketHandler) NotifyBattleFlee(battleID string, fleeUserID, winnerID uint) {
	// 获取用户信息
	fleeUser, _ := h.userService.GetUserByID(fleeUserID)
	winnerUser, _ := h.userService.GetUserByID(winnerID)

	// 通知逃跑者
	if fleeUser != nil {
		fleeMessage := WSMessage{
			Type: "battle_flee",
			Data: map[string]interface{}{
				"message": fmt.Sprintf("你已逃跑，对战失败。当前积分: %d", fleeUser.Rating),
				"result":  "loss",
				"rating":  fleeUser.Rating,
			},
		}
		h.sendToClient(fleeUserID, fleeMessage)
	}

	// 通知获胜者
	if winnerUser != nil {
		winnerMessage := WSMessage{
			Type: "opponent_fled",
			Data: map[string]interface{}{
				"message": fmt.Sprintf("对手已逃跑，你获得胜利！当前积分: %d", winnerUser.Rating),
				"result":  "win",
				"rating":  winnerUser.Rating,
				"opponent": map[string]interface{}{
					"username": fleeUser.Username,
					"rating":   fleeUser.Rating,
				},
			},
		}
		h.sendToClient(winnerID, winnerMessage)
	}

	log.Printf("Notified battle flee for %s: user %d fled, user %d wins", battleID, fleeUserID, winnerID)
}
