package queue

import (
	"context"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestConfiguration(t *testing.T) {
	t.Run("Test", func(t *testing.T) {
		workingDir, err := os.Getwd()
		require.Nil(t, err)

		fileName := workingDir + "/configuration.yaml"
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)

		pipeline, err := BuildPipeline(ctx, fileName)
		require.Nil(t, err)
		go func() {
			err = pipeline.Start(ctx)
			require.Nil(t, err)
		}()

		time.Sleep(time.Second * 5)
		cancel()
	})

}
