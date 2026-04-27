// Copyright (C) 2026 Leazoot
// SPDX-License-Identifier: AGPL-3.0-or-later
// This file is part of ReadBud, licensed under the GNU AGPL v3.
// See LICENSE in the project root or <https://www.gnu.org/licenses/agpl-3.0.html>.

package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

// HashPassword generates a bcrypt hash from a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("crypto.HashPassword: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword compares a plaintext password against a bcrypt hash.
// Returns nil if the password matches.
func VerifyPassword(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("crypto.VerifyPassword: %w", err)
	}
	return nil
}
