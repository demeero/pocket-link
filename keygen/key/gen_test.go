package key

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func TestGenerate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	usedRepo := NewMockUsedKeysRepository(ctrl)
	unusedRepo := NewMockUnusedKeysRepository(ctrl)

	size := 0
	predefinedKeysCount := 20
	step := predefinedKeysCount/10 + 1
	ctx, cancel := context.WithCancel(context.Background())
	unusedRepo.EXPECT().Size(ctx).DoAndReturn(func(context.Context) (int64, error) {
		result := size
		size += step
		if size >= predefinedKeysCount {
			cancel()
		}
		return int64(result), nil
	}).AnyTimes()
	unusedRepo.EXPECT().Store(ctx, gomock.Any()).Return(int64(1), nil).Times(21)
	usedRepo.EXPECT().Exists(ctx, gomock.Any()).Return(false, nil).Times(21)

	Generate(ctx, GeneratorConfig{
		PredefinedKeysCount: uint(predefinedKeysCount),
		Delay:               time.Millisecond * 100,
		KeyLen:              8,
	}, usedRepo, unusedRepo)
}
