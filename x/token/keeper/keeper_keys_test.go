package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenKey_UpperLowercase(t *testing.T) {
	assert.Equal(t, TokenKey("LoV"), TokenKey("lOv"))
	assert.NotEqual(t, TokenKey("LoV1"), TokenKey("lOv"))
}
