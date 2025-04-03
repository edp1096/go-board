// internal/service/auth_service.go
package service

import (
	"context"
	"errors"
	"fmt"
	"go-board/internal/models"
	"go-board/internal/repository"
	"log"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("유효하지 않은 인증 정보")
	ErrUserNotFound       = errors.New("사용자를 찾을 수 없음")
	ErrUserInactive       = errors.New("비활성화된 사용자")
	ErrUsernameTaken      = errors.New("이미 사용 중인 사용자 이름")
	ErrEmailTaken         = errors.New("이미 사용 중인 이메일")
)

type AuthService interface {
	Register(ctx context.Context, username, email, password, fullName string) (*models.User, error)
	Login(ctx context.Context, username, password string) (*models.User, string, error)
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserActiveStatus(ctx context.Context, id int64, active bool) error
	ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error
	AdminChangePassword(ctx context.Context, id int64, newPassword string) error
	DeleteUser(ctx context.Context, id int64) error
	ListUsers(ctx context.Context, offset, limit int, search string) ([]*models.User, int, error)
}

type authService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	tokenExpiry time.Duration
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	// 환경 변수에서 JWT 시크릿 키 로드
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// 개발 환경에서만 사용하는 대체 시크릿을 환경 검사 후 설정
		env := os.Getenv("APP_ENV")
		if env != "production" {
			// 개발 환경에서는 날짜 기반 임의 키 사용
			jwtSecret = "dev_jwt_secret_" + time.Now().Format("20060102")
			log.Println("경고: JWT_SECRET 환경 변수가 설정되지 않았습니다. 임시 시크릿을 생성합니다.")
		} else {
			log.Fatal("운영 환경에서 JWT_SECRET이 설정되지 않았습니다. 애플리케이션을 종료합니다.")
		}
	}

	return &authService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		tokenExpiry: 24 * time.Hour, // 토큰 만료 기간 (예: 24시간)
	}
}

func (s *authService) Register(ctx context.Context, username, email, password, fullName string) (*models.User, error) {
	// 이미 존재하는 사용자 이름인지 확인
	existingUser, err := s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, ErrUsernameTaken
	}

	// 이미 존재하는 이메일인지 확인
	existingUser, err = s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailTaken
	}

	// 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("비밀번호 해싱 오류: %w", err)
	}

	// 사용자 생성
	user := &models.User{
		Username:  username,
		Email:     email,
		Password:  string(hashedPassword),
		FullName:  fullName,
		Role:      models.RoleUser, // 기본 역할은 일반 사용자
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 데이터베이스에 저장
	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("사용자 등록 오류: %w", err)
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.User, string, error) {
	// 사용자 조회
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// 비활성 사용자 확인
	if !user.Active {
		return nil, "", ErrUserInactive
	}

	// 비밀번호 검증
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// JWT 토큰 생성
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("토큰 생성 오류: %w", err)
	}

	return user, token, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*models.User, error) {
	// 토큰 파싱
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// 서명 알고리즘 확인
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("예상치 못한 서명 방법: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("토큰 검증 오류: %w", err)
	}

	// 토큰 클레임 검증
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := int64(claims["user_id"].(float64))

		// 사용자 조회
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, ErrUserNotFound
		}

		// 비활성 사용자 확인
		if !user.Active {
			return nil, ErrUserInactive
		}

		return user, nil
	}

	return nil, errors.New("유효하지 않은 토큰")
}

func (s *authService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *authService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *authService) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// UpdateUserActiveStatus는 사용자의 활성 상태만 업데이트합니다
func (s *authService) UpdateUserActiveStatus(ctx context.Context, id int64, active bool) error {
	return s.userRepo.UpdateActiveStatus(ctx, id, active)
}

func (s *authService) ChangePassword(ctx context.Context, id int64, currentPassword, newPassword string) error {
	// 사용자 조회
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// 현재 비밀번호 검증
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// 새 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("비밀번호 해싱 오류: %w", err)
	}

	// 비밀번호 업데이트
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// AdminChangePassword 관리자가 사용자 비밀번호 변경 (현재 비밀번호 검증 없음)
func (s *authService) AdminChangePassword(ctx context.Context, id int64, newPassword string) error {
	// 사용자 조회
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// 새 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("비밀번호 해싱 오류: %w", err)
	}

	// 비밀번호 업데이트
	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(ctx, user)
}

// DeleteUser 사용자 삭제
func (s *authService) DeleteUser(ctx context.Context, id int64) error {
	// 사용자 존재 확인
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// 사용자 삭제
	return s.userRepo.Delete(ctx, id)
}

// ListUsers 사용자 목록 조회
func (s *authService) ListUsers(ctx context.Context, offset, limit int, search string) ([]*models.User, int, error) {
	// 검색어가 있는 경우
	if search != "" {
		// 검색 로직을 구현 (사용자명, 이메일로 검색)
		// 여기서는 간단히 Repository의 List를 사용하고 애플리케이션 레벨에서 필터링
		users, _, err := s.userRepo.List(ctx, 0, 1000) // 더 큰 수를 가져와서 필터링
		if err != nil {
			return nil, 0, err
		}

		// 검색어로 필터링
		filteredUsers := make([]*models.User, 0)
		for _, user := range users {
			// 사용자명이나 이메일에 검색어가 포함되어 있으면 추가
			if strings.Contains(strings.ToLower(user.Username), strings.ToLower(search)) ||
				strings.Contains(strings.ToLower(user.Email), strings.ToLower(search)) ||
				strings.Contains(strings.ToLower(user.FullName), strings.ToLower(search)) {
				filteredUsers = append(filteredUsers, user)
			}
		}

		// 결과 슬라이싱 (페이지네이션)
		resultUsers := make([]*models.User, 0)
		count := len(filteredUsers)

		// 오프셋 유효성 검사
		if offset >= count {
			return resultUsers, count, nil
		}

		// 끝 인덱스 계산
		end := min(offset+limit, count)

		// 결과 슬라이싱
		if offset < count {
			resultUsers = filteredUsers[offset:end]
		}

		return resultUsers, count, nil
	}

	// 검색어가 없는 경우 기본 목록 조회
	return s.userRepo.List(ctx, offset, limit)
}

// JWT 토큰 생성 헬퍼 함수
func (s *authService) generateToken(user *models.User) (string, error) {
	// 클레임 생성
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(s.tokenExpiry).Unix(),
	}

	// 토큰 생성
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 토큰 서명
	return token.SignedString([]byte(s.jwtSecret))
}
