package services_test

import (
	"io"
	"os"
	"testing"

	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/services"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStoreWithExpectedString() (*storage.MemStorage, string) {
	store := storage.NewMemStorage()
	store.Counter = map[string]int64{
		"test":  1,
		"test1": 2,
	}
	store.Gauge = map[string]float64{
		"test":  1.5,
		"test2": 35.676,
	}

	return store, `{"counter": {"test": 1, "test1": 2}, "gauge": {"test": 1.5, "test2": 35.676}}`
}

func TestUploadService_Save(t *testing.T) {
	logger, err := logger.Initialize("info")
	require.NoError(t, err)

	tests := []struct {
		name  string
		tBody func()
	}{
		{
			name: "success open file",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: file.Name()}
				store := storage.NewMemStorage()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				assert.NotNil(t, uploadService)
			},
		},
		{
			name: "success save metrics in open file",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: file.Name()}
				store, expectedRes := setupStoreWithExpectedString()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				saveErr := uploadService.Save()
				require.NoError(t, saveErr)

				fileBody, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.JSONEq(t, expectedRes, string(fileBody))
			},
		},
		{
			name: "should not save metrics if file path empty string",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				cfg := config.Config{FileStoragePath: ""}
				store, _ := setupStoreWithExpectedString()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				saveErr := uploadService.Save()
				require.NoError(t, saveErr)

				fileBody, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Equal(t, "", string(fileBody))
			},
		},
		{
			name: "should clean file before update",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"test": true}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: file.Name()}

				store, expectedRes := setupStoreWithExpectedString()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				saveErr := uploadService.Save()
				require.NoError(t, saveErr)

				fileBody, err := io.ReadAll(file)
				require.NoError(t, err)

				require.JSONEq(t, expectedRes, string(fileBody))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tBody()
		})
	}
}

func TestUploadService_Load(t *testing.T) {
	logger, err := logger.Initialize("info")
	require.NoError(t, err)

	tests := []struct {
		name  string
		tBody func()
	}{
		{
			name: "should not load storage from file if file path empty string",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"counter": {"test": 1}}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: ""}
				store := storage.NewMemStorage()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				loadErr := uploadService.Load()
				require.NoError(t, loadErr)

				expectedStore := storage.NewMemStorage()

				assert.Equal(t, expectedStore, store)
			},
		},
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

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				loadErr := uploadService.Load()
				require.NoError(t, loadErr)

				expectedStore := storage.NewMemStorage()

				assert.Equal(t, expectedStore, store)
			},
		},
		{
			name: "should success load data from file and update mem storage",
			tBody: func() {
				file, err := os.CreateTemp("./", "*db.json")
				require.NoError(t, err)

				defer os.Remove(file.Name())

				_, writeErr := file.Write([]byte(`{"gauge": {}, "counter": {"test": 1}}`))
				require.NoError(t, writeErr)

				_, seekErr := file.Seek(0, io.SeekStart)
				require.NoError(t, seekErr)

				cfg := config.Config{FileStoragePath: file.Name(), Restore: true}
				store := storage.NewMemStorage()

				uploadService, err := services.NewUploadService(&cfg, store, logger)
				require.NoError(t, err)

				defer uploadService.Close()

				loadErr := uploadService.Load()
				require.NoError(t, loadErr)

				expectedStore := storage.NewMemStorage()
				expectedStore.Counter = map[string]int64{
					"test": 1,
				}

				assert.Equal(t, expectedStore, store)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.tBody()
		})
	}
}
