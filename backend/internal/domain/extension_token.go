// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package domain

import "time"

// ExtensionToken is a long-lived bearer credential issued to the browser extension
// so the plugin can read article packages without going through the interactive
// JWT login flow. One user may have many tokens (e.g., one per browser/device).
//
// Token storage rules:
//   - The plaintext token is shown to the user exactly ONCE (at issuance).
//   - The DB only ever stores `token_hash` (sha256 hex) and a short `token_prefix`
//     for human-friendly identification in the management UI.
//   - Revocation is soft (sets `revoked_at`) so audit history is preserved.
type ExtensionToken struct {
	BaseModel
	UserID      int64      `gorm:"type:bigint;not null;index" json:"user_id"`
	Name        string     `gorm:"type:varchar(64);not null" json:"name"`
	TokenHash   string     `gorm:"type:varchar(128);not null;uniqueIndex" json:"-"`
	TokenPrefix string     `gorm:"type:varchar(16);not null" json:"token_prefix"`
	LastUsedAt  *time.Time `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `gorm:"type:timestamptz" json:"expires_at,omitempty"`
	RevokedAt   *time.Time `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
}

// TableName overrides the default table name.
func (ExtensionToken) TableName() string {
	return "extension_tokens"
}
