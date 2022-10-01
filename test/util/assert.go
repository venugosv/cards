package util

import (
	"fmt"
	"testing"

	"github.com/anz-bank/equals"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func CheckTestAndAuditLogs(t *testing.T, got interface{}, want interface{}, wantErr error, err error, b fmt.Stringer) {
	if wantErr != nil {
		require.Error(t, err)
		assert.Contains(t, err.Error(), wantErr.Error())
		assert.NotContains(t, b.String(), "no error sent to audit log")
	} else {
		require.NoError(t, err)
		equals.AssertJson(t, want, got)
		assert.NotContains(t, b.String(), "no response data sent to audit log")
	}
}
