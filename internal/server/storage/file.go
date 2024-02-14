package storage

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/sodiqit/metricpulse.git/internal/entities"
	"github.com/sodiqit/metricpulse.git/internal/logger"
	"github.com/sodiqit/metricpulse.git/internal/server/config"
)

type FileStorage struct {
	cfg     *config.Config
	storage *MemStorage
	file    *os.File
	logger  logger.ILogger
}

func (s *FileStorage) SaveGaugeMetric(metricType string, value float64) (float64, error) {
	res, err := s.storage.SaveGaugeMetric(metricType, value)

	if s.cfg.StoreInterval != 0 || err != nil {
		return res, err
	}

	err = s.save()

	return res, err
}

func (s *FileStorage) SaveCounterMetric(metricType string, value int64) (int64, error) {
	res, err := s.storage.SaveCounterMetric(metricType, value)

	if s.cfg.StoreInterval != 0 || err != nil {
		return res, err
	}

	err = s.save()

	return res, err
}

func (s *FileStorage) GetGaugeMetric(metricName string) (float64, error) {
	return s.storage.GetGaugeMetric(metricName)
}

func (s *FileStorage) GetCounterMetric(metricName string) (int64, error) {
	return s.storage.GetCounterMetric(metricName)
}

func (s *FileStorage) GetAllMetrics() (entities.TotalMetrics, error) {
	return s.storage.GetAllMetrics()
}

func (s *FileStorage) Init(ctx context.Context) error {
	if s.cfg.FileStoragePath == "" {
		return errors.New("file not provided for start file storage")
	}

	file, err := os.OpenFile(s.cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0666)

	if err != nil {
		return err
	}

	s.file = file

	err = s.storeInterval()

	if err != nil {
		return err
	}

	err = s.load()

	return err
}

func (s *FileStorage) Ping(ctx context.Context) error {
	return nil
}

func (s *FileStorage) Close(ctx context.Context) error {
	if s.file == nil {
		return nil
	}

	defer s.file.Close()

	err := s.save()

	return err
}

func (u *FileStorage) load() error {
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

func (u *FileStorage) save() error {
	if u.file == nil {
		return errors.New("file not found")
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

func (u *FileStorage) storeInterval() error {
	if u.cfg.StoreInterval == 0 {
		return nil
	}
	storeDuration := time.Duration(u.cfg.StoreInterval) * time.Second
	go func() {
		for {
			time.Sleep(storeDuration)
			err := u.save()

			if err != nil {
				u.logger.Errorw("error while saving", "error", err)
			}
		}
	}()
	return nil //TODO: fix this
}

func (u *FileStorage) cleanFile() error {
	_, err := u.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = u.file.Truncate(0)
	return err
}

func NewFileStorage(cfg *config.Config, storage *MemStorage, logger logger.ILogger) *FileStorage {
	return &FileStorage{cfg, storage, nil, logger}
}
