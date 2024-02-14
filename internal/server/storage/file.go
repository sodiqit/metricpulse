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

func (s *FileStorage) SaveGaugeMetric(ctx context.Context, metricType string, value float64) (float64, error) {
	res, err := s.storage.SaveGaugeMetric(ctx, metricType, value)

	if s.cfg.StoreInterval != 0 || err != nil {
		return res, err
	}

	err = s.save(ctx)

	return res, err
}

func (s *FileStorage) SaveCounterMetric(ctx context.Context, metricType string, value int64) (int64, error) {
	res, err := s.storage.SaveCounterMetric(ctx, metricType, value)

	if s.cfg.StoreInterval != 0 || err != nil {
		return res, err
	}

	err = s.save(ctx)

	return res, err
}

func (s *FileStorage) GetGaugeMetric(ctx context.Context, metricName string) (float64, error) {
	return s.storage.GetGaugeMetric(ctx, metricName)
}

func (s *FileStorage) GetCounterMetric(ctx context.Context, metricName string) (int64, error) {
	return s.storage.GetCounterMetric(ctx, metricName)
}

func (s *FileStorage) GetAllMetrics(ctx context.Context) (entities.TotalMetrics, error) {
	return s.storage.GetAllMetrics(ctx)
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

	err = s.storeInterval(ctx)

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

	err := s.save(ctx)

	return err
}

func (s *FileStorage) load() error {
	if s.file == nil || !s.cfg.Restore {
		return nil
	}

	data, err := io.ReadAll(s.file)

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

	initMetricsErr := s.storage.InitMetrics(metrics)

	if initMetricsErr != nil {
		return initMetricsErr
	}

	s.logger.Infow("success load metrics", "metrics", metrics, "filePath", s.cfg.FileStoragePath)

	return nil
}

func (s *FileStorage) save(ctx context.Context) error {
	if s.file == nil {
		return errors.New("file not found")
	}

	metrics, err := s.storage.GetAllMetrics(ctx)

	if err != nil {
		return err
	}

	res, err := json.Marshal(metrics)

	if err != nil {
		return err
	}

	cleanErr := s.cleanFile()

	if cleanErr != nil {
		return cleanErr
	}

	_, writeErr := s.file.Write(res)

	if writeErr != nil {
		return writeErr
	}

	s.logger.Infow("success save in file", "metrics", string(res), "filePath", s.cfg.FileStoragePath)

	return nil
}

func (s *FileStorage) storeInterval(ctx context.Context) error {
	if s.cfg.StoreInterval == 0 {
		return nil
	}
	storeDuration := time.Duration(s.cfg.StoreInterval) * time.Second
	go func() {
		for {
			time.Sleep(storeDuration)
			err := s.save(ctx)

			if err != nil {
				s.logger.Errorw("error while saving", "error", err)
			}
		}
	}()
	return nil //TODO: fix this
}

func (s *FileStorage) cleanFile() error {
	_, err := s.file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	err = s.file.Truncate(0)
	return err
}

func NewFileStorage(cfg *config.Config, storage *MemStorage, logger logger.ILogger) *FileStorage {
	return &FileStorage{cfg, storage, nil, logger}
}
