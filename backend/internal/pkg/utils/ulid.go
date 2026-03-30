package utils

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// NewULID generates a new ULID string for use as public_id.
func NewULID() string {
	return ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String()
}
