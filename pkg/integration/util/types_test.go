package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypes(t *testing.T) {
	t.Run("GetInt64Val", func(t *testing.T) {
		assert.Equal(t, GetInt64Val(ToInt64Ptr(10)), int64(10))
		assert.Equal(t, GetInt64Val(nil), int64(0))
	})
	t.Run("GetBoolVal", func(t *testing.T) {
		assert.Equal(t, GetBoolVal(ToBoolPtr(true)), true)
		assert.Equal(t, GetBoolVal(nil), false)
	})
	t.Run("GetStringVal", func(t *testing.T) {
		assert.Equal(t, GetStringVal(ToStringPtr("test")), "test")
		assert.Equal(t, GetStringVal(nil), "")
	})
}
