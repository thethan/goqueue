package queue

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConfiguration(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		if testing.Short() {
			t.Skipf("skpping integration test...")
		}
		workingDir, err := os.Getwd()
		require.Nil(t, err)

		fileName := workingDir + "/configuration.yaml"
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		pipeline, err := BuildPipeline(ctx, fileName)
		require.Nil(t, err)
		//go func() {
		err = pipeline.Start(ctx)
		require.Nil(t, err)
		//}()

		//time.Sleep(time.Second * 5)
		_ = cancel
		//cancel()
	})
	t.Run("LRange Test", func(t *testing.T) {
		if testing.Short() {
			t.Skipf("skpping integration test...")
		}
		workingDir, err := os.Getwd()
		require.Nil(t, err)

		fileName := workingDir + "/configuration-lrange.yaml"
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		pipeline, err := BuildPipeline(ctx, fileName)
		require.Nil(t, err)
		//go func() {
		err = pipeline.Start(ctx)
		require.Nil(t, err)
		//}()

		//time.Sleep(time.Second * 5)
		_ = cancel
		//cancel()
	})
}
