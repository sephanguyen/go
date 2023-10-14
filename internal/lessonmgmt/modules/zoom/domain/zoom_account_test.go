package domain

import (
	"context"
	"testing"
	"time"

	"github.com/tkuchiki/faketime"
	"gotest.tools/assert"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func TestZoomAccountBuilder(t *testing.T) {
	t.Run("should create zoom account success", func(t *testing.T) {
		now := time.Now()
		f := faketime.NewFaketimeWithTime(now)
		defer f.Undo()
		f.Do()

		zoomAccount, _ := NewZoomAccountBuilder().
			WithEmail("test@gmail.com").
			WithAction("Upsert").
			WithID("id").
			WithUsername("username").Build()
		assert.Equal(t, now, *zoomAccount.CreatedAt)
		assert.Equal(t, now, *zoomAccount.UpdatedAt)
		assert.Equal(t, "test@gmail.com", zoomAccount.Email)
		assert.Equal(t, ZoomAction("Upsert"), zoomAccount.Action)
		assert.Equal(t, "id", zoomAccount.ID)
		assert.Equal(t, "username", zoomAccount.UserName)
	})
	t.Run("should throw error if empty email", func(t *testing.T) {
		now := time.Now()
		f := faketime.NewFaketimeWithTime(now)
		defer f.Undo()
		f.Do()

		_, err := NewZoomAccountBuilder().
			WithAction("Upsert").
			WithID("id").
			WithUsername("username").Build()
		assert.Equal(t, "invalid zoom account detail: email could not be empty", err.Error())

	})
	t.Run("should throw error if empty email", func(t *testing.T) {
		now := time.Now()
		f := faketime.NewFaketimeWithTime(now)
		defer f.Undo()
		f.Do()

		_, err := NewZoomAccountBuilder().
			WithEmail("test@gmail.com").
			WithID("id").
			WithUsername("username").Build()
		assert.Equal(t, "invalid zoom account detail: action could not be empty", err.Error())

	})

	t.Run("should create zoom account success when action is delete", func(t *testing.T) {
		now := time.Now()
		f := faketime.NewFaketimeWithTime(now)
		defer f.Undo()
		f.Do()

		zoomAccount, _ := NewZoomAccountBuilder().
			WithEmail("test@gmail.com").
			WithAction("Delete").
			WithID("id").
			WithUsername("username").Build()
		assert.Equal(t, now, *zoomAccount.CreatedAt)
		assert.Equal(t, now, *zoomAccount.UpdatedAt)
		assert.Equal(t, now, *zoomAccount.DeletedAt)
		assert.Equal(t, "test@gmail.com", zoomAccount.Email)
		assert.Equal(t, ZoomAction("Delete"), zoomAccount.Action)
		assert.Equal(t, "id", zoomAccount.ID)
		assert.Equal(t, "username", zoomAccount.UserName)
	})
}
