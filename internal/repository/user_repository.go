// internal/repository/user_repository.go
package repository

import (
	"context"
	"time"

	"github.com/edp1096/go-board/internal/models"

	"github.com/uptrace/bun"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	UpdateActiveStatus(ctx context.Context, id int64, active bool) error
	UpdateApprovalStatus(ctx context.Context, id int64, status models.ApprovalStatus) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, offset, limit int) ([]*models.User, int, error)
	SearchUsers(ctx context.Context, query string, offset, limit int) ([]*models.User, error)
	GetPendingApprovalUsers(ctx context.Context) ([]*models.User, error)

	GetByExternalID(ctx context.Context, externalID, externalSystem string) (*models.User, error)
	InvalidateTokens(ctx context.Context, userID int64) error
}

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	_, err := r.db.NewInsert().Model(user).Exec(ctx)
	return err
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("username = ?", username).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	_, err := r.db.NewUpdate().Model(user).WherePK().Exec(ctx)
	return err
}

func (r *userRepository) UpdateActiveStatus(ctx context.Context, id int64, active bool) error {
	_, err := r.db.NewUpdate().
		Table("users").
		Set("active = ?", active).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *userRepository) UpdateApprovalStatus(ctx context.Context, id int64, status models.ApprovalStatus) error {
	_, err := r.db.NewUpdate().
		Table("users").
		Set("approval_status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.NewDelete().Model((*models.User)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*models.User, int, error) {
	var users []*models.User
	count, err := r.db.NewSelect().Model(&users).Limit(limit).Offset(offset).ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}
	return users, count, nil
}

// SearchUsers - 사용자 검색
func (r *userRepository) SearchUsers(ctx context.Context, query string, offset, limit int) ([]*models.User, error) {
	if limit <= 0 {
		limit = 10
	}

	// 쿼리 작성 - 사용자명, 이메일, 이름에서 검색
	selectQuery := r.db.NewSelect().
		Model((*models.User)(nil)).
		Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?",
						"%"+query+"%", "%"+query+"%", "%"+query+"%").
		Where("active = ?", true). // 활성 사용자만 검색
		Limit(limit).
		Offset(offset)

	var users []*models.User
	err := selectQuery.Scan(ctx, &users)

	return users, err
}

// GetPendingApprovalUsers - 승인 대기 중인 사용자 목록 조회
func (r *userRepository) GetPendingApprovalUsers(ctx context.Context) ([]*models.User, error) {
	var users []*models.User
	err := r.db.NewSelect().
		Model(&users).
		Where("approval_status = ?", models.ApprovalPending).
		OrderExpr("created_at ASC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return users, nil
}

// GetByExternalID 구현
func (r *userRepository) GetByExternalID(ctx context.Context, externalID, externalSystem string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("external_id = ? AND external_system = ?", externalID, externalSystem).
		Scan(ctx)

	if err != nil {
		return nil, err
	}
	return user, nil
}

// InvalidateTokens 구현
func (r *userRepository) InvalidateTokens(ctx context.Context, userID int64) error {
	_, err := r.db.NewUpdate().
		Table("users").
		Set("token_invalidated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(ctx)

	return err
}
