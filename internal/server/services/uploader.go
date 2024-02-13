package services

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
	"github.com/sodiqit/metricpulse.git/internal/server/storage"
)

type IUploadService interface {
	Save() error
	Load() error
	StoreInterval() error
	Close() error
}

type UploadService struct {
	cfg     *config.Config
	storage storage.IStorage
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

	var metrics entities.TotalMetrics

	jsonErr := json.Unmarshal(data, &metrics)

	if jsonErr != nil {
		return jsonErr
	}

	initMetricsErr := u.storage.InitMetrics(metrics)

	if initMetricsErr != nil {
		return initMetricsErr
	}

	u.logger.Infow("success load metrics", "metrics", metrics, "filePath", u.cfg.FileStoragePath)

	return nil
}

func (u *UploadService) Save() error {
	if u.file == nil {
		return nil
	}

	metrics, err := u.storage.GetAllMetrics()

	if err != nil {
		return err
	}

	res, err := json.Marshal(metrics)

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

func (u *UploadService) StoreInterval() error {
	if u.cfg.StoreInterval == 0 {
		return nil
	}
	storeDuration := time.Duration(u.cfg.StoreInterval) * time.Second
	go func() {
		for {
			time.Sleep(storeDuration)
			err := u.Save()

			if err != nil {
				u.logger.Errorw("error while saving", "error", err)
			}
		}
	}()
	return nil //TODO: fix this
}

func (u *UploadService) cleanFile() error {
	_, err := u.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = u.file.Truncate(0)
	return err
}

func NewUploadService(cfg *config.Config, storage storage.IStorage, logger logger.ILogger) (*UploadService, error) {
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
