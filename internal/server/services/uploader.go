package services

import (
	"encoding/json"
	"io"
	"os"

	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IUploadService interface {
	Save() error
	Load() error
	Close() error
}

type UploadService struct {
	cfg     *config.Config
	storage *storage.MemStorage
	file    *os.File
	logger  logger.ILogger
}

func (u *UploadService) Close() error {
	if u.file == nil {
		return nil
	}
	return u.file.Close()
}

func (u *UploadService) Load() error {
	if u.file == nil || !u.cfg.Restore {
		return nil
	}

	data, err := io.ReadAll(u.file)

	if err != nil {
		return err
	}

	if string(data) == "" {
		return nil
	}

	var storage storage.MemStorage

	jsonErr := json.Unmarshal(data, &storage)

	if jsonErr != nil {
		return jsonErr
	}

	u.storage.Counter = storage.Counter
	u.storage.Gauge = storage.Gauge

	u.logger.Infow("success load metrics", "metrics", u.storage, "filePath", u.cfg.FileStoragePath)

	return nil
}

func (u *UploadService) Save() error {
	if u.file == nil {
		return nil
	}

	res, err := json.Marshal(u.storage)

	if err != nil {
		return err
	}

	cleanErr := u.cleanFile()

	if cleanErr != nil {
		return cleanErr
	}

	_, writeErr := u.file.Write(res)

	if writeErr != nil {
		return writeErr
	}

	u.logger.Infow("success save in file", "metrics", string(res), "filePath", u.cfg.FileStoragePath)

	return nil
}

func (u *UploadService) cleanFile() error {
	_, err := u.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = u.file.Truncate(0)
	return err
}

func NewUploadService(cfg *config.Config, storage *storage.MemStorage, logger logger.ILogger) (*UploadService, error) {
	var file *os.File

	if cfg.FileStoragePath != "" {
		f, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)

		if err != nil {
			return nil, err
		}

		file = f
	}

	return &UploadService{cfg, storage, file, logger}, nil
}
