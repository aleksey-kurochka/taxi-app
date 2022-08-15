package rnd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRandomInt(t *testing.T) {
	r1 := RandomInt(100)
	r2 := RandomInt(100)

	assert.NotEqual(t, r1, r2)
}

func TestRandomStr(t *testing.T) {
	s := RandomStr(5)

	assert.Equal(t, 5, len(s))
}
