package metricuploader

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

type Uploader interface {
	Save() error
	Load() error
	StoreInterval() error
	Close() error
}

type MetricUploader struct {
	cfg     *config.Config
	storage storage.Storage
	file    *os.File
	logger  logger.ILogger
}

func (u *MetricUploader) Close() error {
	if u.file == nil {
		return nil
	}
	return u.file.Close()
}

func (u *MetricUploader) Load() error {
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

func (u *MetricUploader) Save() error {
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

func (u *MetricUploader) StoreInterval() error {
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

func (u *MetricUploader) cleanFile() error {
	_, err := u.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = u.file.Truncate(0)
	return err
}

func New(cfg *config.Config, storage storage.Storage, logger logger.ILogger) (*MetricUploader, error) {
	var file *os.File

	if cfg.FileStoragePath != "" {
		f, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)

		if err != nil {
			return nil, err
		}

		file = f
	}

	return &MetricUploader{cfg, storage, file, logger}, nil
}
