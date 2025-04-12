// internal/models/user.go
package models

import (
	"time"

	"github.com/uptrace/bun"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID             int64          `bun:"id,pk,autoincrement" json:"id"`
	Username       string         `bun:"username,unique,notnull" json:"username"`
	Email          string         `bun:"email,unique,notnull" json:"email"`
	Password       string         `bun:"password,notnull" json:"-"`
	FullName       string         `bun:"full_name" json:"fullName"`
	Role           Role           `bun:"role,notnull" json:"role"`
	Active         bool           `bun:"active,notnull,default:true" json:"active"`
	ApprovalStatus ApprovalStatus `bun:"approval_status,notnull,default:'pending'" json:"approvalStatus"`
	ApprovalDue    *time.Time     `bun:"approval_due" json:"approvalDue"`
	CreatedAt      time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"createdAt"`
	UpdatedAt      time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updatedAt"`
}
