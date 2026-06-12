package ids

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

var entropy = ulid.Monotonic(rand.Reader, 0)

func New() string {
	return ulid.MustNew(
		ulid.Timestamp(time.Now()),
		entropy,
	).String()
}

func NewJobID() string {
	return New()
}

func NewAttemptID() string {
	return New()
}

func NewOutboxID() string {
	return New()
}
