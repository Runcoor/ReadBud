// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package domain

// User represents the users table per spec Section 11.1.
type User struct {
	BaseModel
	Username     string `gorm:"type:varchar(64);not null;uniqueIndex" json:"username"`
	PasswordHash string `gorm:"type:varchar(255);not null" json:"-"`
	Nickname     string `gorm:"type:varchar(64);not null" json:"nickname"`
	Role         string `gorm:"type:varchar(32);not null;default:'editor'" json:"role"`
	Status       int16  `gorm:"type:smallint;not null;default:1" json:"status"`
}

// TableName overrides the default table name.
func (User) TableName() string {
	return "users"
}

// User role constants.
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
)
