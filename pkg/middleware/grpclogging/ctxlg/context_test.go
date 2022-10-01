package ctxlg

import (
	"context"
	"testing"

	grpcctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"

	"github.com/stretchr/testify/assert"
)

func TestMergeFields(t *testing.T) {
	t.Run("", func(t *testing.T) {
		m1 := map[string]interface{}{
			"key01": "value01",
		}
		m2 := map[string]interface{}{
			"key02": "value02",
		}
		want := map[string]interface{}{
			"key01": "value01",
			"key02": "value02",
		}
		got := MergeFields(m1, m2)
		assert.Equal(t, want, got)
	})
}

func TestAddFields(t *testing.T) {
	t.Run("", func(t *testing.T) {
		fields := make(map[string]interface{})
		ctx := ToContext(context.Background(), &CtxLogger{Fields: fields})

		add := map[string]interface{}{
			"key01": "value01",
		}
		AddFields(ctx, add)

		got := Extract(ctx)

		assert.Equal(t, add, got.Fields)
	})
	t.Run("", func(t *testing.T) {
		ctx := context.Background()

		add := map[string]interface{}{
			"key01": "value01",
		}
		AddFields(ctx, add)

		assert.Equal(t, context.Background(), ctx)
	})
}

func TestExtract(t *testing.T) {
	t.Run("", func(t *testing.T) {
		fields := make(map[string]interface{})
		ctxLogger := &CtxLogger{Fields: fields}
		ctx := ToContext(context.Background(), ctxLogger)

		got := Extract(ctx)

		assert.Equal(t, ctxLogger, got)
	})
	t.Run("", func(t *testing.T) {
		got := Extract(context.Background())

		assert.Equal(t, &CtxLogger{make(map[string]interface{})}, got)
	})
}

// This test was taken from fabric-selfservice
func TestTagsToFields(t *testing.T) {
	t.Run("", func(t *testing.T) {
		// given:
		fields := map[string]interface{}{
			"key01": "value01",
		} // added fields only to make sure it doesn't affect the tags
		tagsMap := map[string]interface{}{
			"key02": "value02",
			"key03": []map[string]interface{}{
				{
					"key04": "value04",
				},
			},
		}
		ctxLogger := &CtxLogger{Fields: fields}
		ctx := ToContext(context.Background(), ctxLogger)

		tags := grpcctxtags.NewTags()
		for i, v := range tagsMap {
			tags.Set(i, v)
		}

		ctx = grpcctxtags.SetInContext(ctx, tags)

		// when:
		got := TagsToFields(ctx)

		// then:
		assert.Equal(t, tagsMap, got)
	})
}
