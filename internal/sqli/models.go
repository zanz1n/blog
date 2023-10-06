// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0

package sqli

import (
	"database/sql/driver"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

type UserRole string

const (
	UserRoleADMIN     UserRole = "ADMIN"
	UserRolePUBLISHER UserRole = "PUBLISHER"
	UserRoleCOMMON    UserRole = "COMMON"
)

func (e *UserRole) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = UserRole(s)
	case string:
		*e = UserRole(s)
	default:
		return fmt.Errorf("unsupported scan type for UserRole: %T", src)
	}
	return nil
}

type NullUserRole struct {
	UserRole UserRole
	Valid    bool // Valid is true if UserRole is not NULL
}

// Scan implements the Scanner interface.
func (ns *NullUserRole) Scan(value interface{}) error {
	if value == nil {
		ns.UserRole, ns.Valid = "", false
		return nil
	}
	ns.Valid = true
	return ns.UserRole.Scan(value)
}

// Value implements the driver Valuer interface.
func (ns NullUserRole) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return string(ns.UserRole), nil
}

type Post struct {
	ID          pgtype.UUID
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
	Title       string
	Content     []byte
	Topics      []byte
	Description string
	ThumbImage  pgtype.UUID
	UserID      pgtype.UUID
}

type User struct {
	ID        pgtype.UUID
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
	Email     string
	Username  string
	Password  string
	Role      UserRole
}
