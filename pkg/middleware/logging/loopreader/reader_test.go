package loopreader

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestNewReader(t *testing.T) {
	want := []byte("oprah winfrey")
	in := strings.NewReader(string(want))
	reader, err := New(in)
	require.NoError(t, err)

	bytes, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, bytes, want)
	require.NoError(t, reader.Close())

	bytes, err = io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, bytes, want)
	require.NoError(t, reader.Close())
}
