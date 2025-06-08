package services

import (
	"context"
	"crypto/md5"
	"fmt"
	"go-saber-system/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type UserService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUserService(db *gorm.DB, redis *redis.Client) *UserService {
	return &UserService{
		db:    db,
		redis: redis,
	}
}

// Register 用户注册
func (s *UserService) Register(username, email, password string) (*models.User, error) {
	hashedPassword := s.hashPassword(password)

	user := &models.User{
		Username: username,
		Email:    email,
		Password: hashedPassword,
		Rating:   1000, // 初始积分
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (*models.User, error) {
	hashedPassword := s.hashPassword(password)

	var user models.User
	err := s.db.Where("username = ? AND password = ?", username, hashedPassword).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateRating 更新用户积分
func (s *UserService) UpdateRating(userID uint, newRating int, isWin bool) error {
	updates := map[string]interface{}{
		"rating": newRating,
	}

	if isWin {
		updates["wins"] = gorm.Expr("wins + ?", 1)
	} else {
		updates["losses"] = gorm.Expr("losses + ?", 1)
	}

	return s.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

// GetUserStatus 获取用户在线状态
func (s *UserService) GetUserStatus(userID uint) (string, error) {
	ctx := context.Background()
	status, err := s.redis.Get(ctx, fmt.Sprintf("user_status:%d", userID)).Result()
	if err == redis.Nil {
		return "offline", nil
	}
	return status, err
}

// SetUserStatus 设置用户在线状态
func (s *UserService) SetUserStatus(userID uint, status string) error {
	ctx := context.Background()
	return s.redis.Set(ctx, fmt.Sprintf("user_status:%d", userID), status, 30*time.Minute).Err()
}

// GetLeaderboard 获取排行榜
func (s *UserService) GetLeaderboard(limit int) ([]models.User, error) {
	var users []models.User
	err := s.db.Order("rating DESC").Limit(limit).Find(&users).Error
	return users, err
}

// IsUserInMatch 检查用户是否在匹配中
func (s *UserService) IsUserInMatch(userID uint) (bool, error) {
	ctx := context.Background()
	exists, err := s.redis.Exists(ctx, fmt.Sprintf("match_queue:%d", userID)).Result()
	return exists > 0, err
}

// hashPassword 密码哈希
func (s *UserService) hashPassword(password string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

// CalculateNewRating ELO积分计算
func (s *UserService) CalculateNewRating(winnerRating, loserRating int) (int, int) {
	k := 32.0 // K因子

	// 计算期望值
	expectedWin := 1.0 / (1.0 + pow(10, float64(loserRating-winnerRating)/400.0))
	expectedLose := 1.0 / (1.0 + pow(10, float64(winnerRating-loserRating)/400.0))

	// 计算新积分
	newWinnerRating := winnerRating + int(k*(1.0-expectedWin))
	newLoserRating := loserRating + int(k*(0.0-expectedLose))

	return newWinnerRating, newLoserRating
}

// pow 简单的幂函数实现
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}
