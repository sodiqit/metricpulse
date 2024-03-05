//go:build !race

package storage_test

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/sodiqit/metricpulse.git/pkg/retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFile(t *testing.T, file *os.File) string {
	fileBody, err := io.ReadAll(file)
	require.NoError(t, err)

	return string(fileBody)
}

func TestFileStorage_SaveInFile(t *testing.T) {
	logger, err := logger.Initialize("info")
	require.NoError(t, err)
	ctx := context.Background()

	tests := []struct {
		name  string
		tBody func()
	}{
		{
			name: "should success open file",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: file.Name()}
				store := storage.NewMemStorage()

				fileStorage := storage.NewFileStorage(&cfg, store, logger)
				defer fileStorage.Close(ctx)

				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)
			},
		},
		{
			name: "should sync save in file if store interval == 0",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: file.Name()}

				expectedRes := `{"counter": {"test": 1}, "gauge": {}}`

				fileStorage := storage.NewFileStorage(&cfg, storage.NewMemStorage(), logger)
				defer fileStorage.Close(ctx)
				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)

				_, err = fileStorage.SaveCounterMetric(ctx, "test", 1)
				require.NoError(t, err)

				res := readFile(t, file)
				assert.JSONEq(t, expectedRes, res)
			},
		},
		{
			name: "should async save in file if store interval > 0",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: file.Name(), StoreInterval: 1}

				expectedRes := `{"counter": {"test": 1}, "gauge": {}}`

				fileStorage := storage.NewFileStorage(&cfg, storage.NewMemStorage(), logger)
				defer fileStorage.Close(ctx)
				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)

				_, err = fileStorage.SaveCounterMetric(ctx, "test", 1)
				require.NoError(t, err)

				//read file before async update
				res := readFile(t, file)
				assert.Equal(t, "", res)

				sleepDur := time.Duration(cfg.StoreInterval)*time.Second + time.Duration(100)*time.Millisecond

				time.Sleep(sleepDur)

				//read file before async update
				res = readFile(t, file)
				assert.JSONEq(t, expectedRes, res)
			},
		},
		{
			name: "should clean file before update",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"test123": true}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: file.Name()}

				expectedRes := `{"counter": {"test": 1}, "gauge": {}}`

				fileStorage := storage.NewFileStorage(&cfg, storage.NewMemStorage(), logger)
				defer fileStorage.Close(ctx)
				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)

				_, err = fileStorage.SaveCounterMetric(ctx, "test", 1)
				require.NoError(t, err)

				res := readFile(t, file)
				assert.JSONEq(t, expectedRes, res)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tBody()
		})
	}
}

func TestFileStorage_LoadFromFile(t *testing.T) {
	logger, err := logger.Initialize("info")
	require.NoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name  string
		tBody func()
	}{
		{
			name: "should not load storage from file if restore option is false",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"counter": {"test": 1}}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: file.Name(), Restore: false}
				store := storage.NewMemStorage()

				fileStorage := storage.NewFileStorage(&cfg, store, logger)
				defer fileStorage.Close(ctx)
				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)

				store1 := storage.NewMemStorage()

				expectedMetrics, err := store1.GetAllMetrics(ctx)
				require.NoError(t, err)

				resultMetrics, err := store.GetAllMetrics(ctx)
				require.NoError(t, err)

				assert.Equal(t, expectedMetrics, resultMetrics)
			},
		},
		{
			name: "should success load data from file and update mem storage",
			tBody: func() {

				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"counter": {"test": 1}}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: file.Name(), Restore: true}
				store := storage.NewMemStorage()

				fileStorage := storage.NewFileStorage(&cfg, store, logger)
				defer fileStorage.Close(ctx)
				err = fileStorage.Init(ctx, retry.EmptyBackoff)
				require.NoError(t, err)

				expectedMetrics := entities.TotalMetrics{Counter: map[string]int64{
					"test": 1,
				}}

				resultMetrics, err := fileStorage.GetAllMetrics(ctx)
				require.NoError(t, err)

				assert.Equal(t, expectedMetrics, resultMetrics)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tBody()
		})
	}
}
