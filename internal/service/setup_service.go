// internal/service/setup_service.go
package service

import (
	"context"
	"go-board/internal/models"
	"go-board/internal/repository"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SetupService 인터페이스
type SetupService interface {
	IsAdminExists(ctx context.Context) (bool, error)
	CreateAdminUser(ctx context.Context, username, email, password, fullName string) (*models.User, error)
}

// setupService 구현체
type setupService struct {
	userRepo repository.UserRepository
}

// NewSetupService는 새로운 SetupService 인스턴스를 생성합니다
func NewSetupService(userRepo repository.UserRepository) SetupService {
	return &setupService{
		userRepo: userRepo,
	}
}

// IsAdminExists는 관리자 계정이 이미 존재하는지 확인합니다
func (s *setupService) IsAdminExists(ctx context.Context) (bool, error) {
	// 관리자 계정을 찾기 위한 검색
	users, err := s.userRepo.SearchUsers(ctx, "", 0, 1)
	if err != nil {
		return false, err
	}

	// 모든 사용자를 검색하여 관리자가 있는지 확인
	if len(users) > 0 {
		for _, user := range users {
			if user.Role == models.RoleAdmin {
				return true, nil
			}
		}
	}

	// 사용자가 없거나 관리자가 없는 경우
	return false, nil
}

// CreateAdminUser는 새로운 관리자 계정을 생성합니다
func (s *setupService) CreateAdminUser(ctx context.Context, username, email, password, fullName string) (*models.User, error) {
	// 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 관리자 사용자 생성
	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		FullName:  fullName,
		Role:      models.RoleAdmin, // 관리자 역할 부여
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 데이터베이스에 저장
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
