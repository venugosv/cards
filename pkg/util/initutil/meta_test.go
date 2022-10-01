package initutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildMetadata_String(t *testing.T) {
	type fields struct {
		ApplicationName string
		Version         string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "successful print",
			fields: fields{
				ApplicationName: "test",
				Version:         "1234",
			},
			want: "app=test version=1234",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			b := &BuildMetadata{
				ApplicationName: tt.fields.ApplicationName,
				Version:         tt.fields.Version,
			}
			got := b.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

var testsNewBuildMetadata = []struct {
	name string
	want *BuildMetadata
}{
	{
		name: "BuildMetadata",
		want: &BuildMetadata{
			ApplicationName: "",
			Version:         "",
		},
	},
}

func TestNewBuildMetadata(t *testing.T) {
	for _, test := range testsNewBuildMetadata {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			got := NewBuildMetadata()
			assert.Equal(t, tt.want, got)
		})
	}
}
