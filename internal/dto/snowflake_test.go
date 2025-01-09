package dto_test

import (
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zanz1n/blog/internal/dto"
)

func TestGenerateSnowflake(t *testing.T) {
	timestamp := time.Now().Round(time.Millisecond)
	rng := rand.Uint32() & dto.SnowflakeRandMask

	snowflake := dto.NewSnowflakeWith(timestamp, rng)

	assert.Equal(t, timestamp, snowflake.Timestamp())
	assert.Equal(t, rng, snowflake.Rand())
}
