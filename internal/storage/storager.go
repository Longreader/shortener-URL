package storage

import (
	"context"
	"os"

	"github.com/Longreader/go-shortener-url.git/config"
	"github.com/Longreader/go-shortener-url.git/internal/repository"
	"github.com/Longreader/go-shortener-url.git/internal/repository/harddrive"
	"github.com/Longreader/go-shortener-url.git/internal/repository/memory"
	"github.com/Longreader/go-shortener-url.git/internal/repository/postgres"
)

// Storager - интерфейс для хранилища.
type Storager interface {
	Set( // Сократить ссылку.
		ctx context.Context, url repository.URL, userID repository.User,
	) (id repository.ID, err error)
	Get( // Получить оригинальную ссылку по ID.
		ctx context.Context, id repository.ID,
	) (url repository.URL, deleted bool, err error)
	Delete( // Удалить указанные ссылки
		ctx context.Context, ids []repository.ID, user repository.User,
	) error
	RunDelete( // Запуск процесса удаления под паттерн FanIn
	)
	GetAllByUser( // Получить все ссылки пользователя.
		ctx context.Context, user repository.User,
	) (links []repository.LinkData, err error)
	Ping( // Проверить соединение с базой данных
		ctx context.Context,
	) (bool, error)
	Close(ctx context.Context) error
}

// StoragerType - int для хранения типа хранилища.
type StoragerType int

// Константы, которые определяют типы StoragerType.
const (
	MemoryStorage StoragerType = iota // Хранилище во временной памяти.
	FileStorage                       // Хранилище в текстовом файле.
	PsqlStorage                       // Хранилище в базе данных Postgres.
)

// NewStorager - конструктор для хранилища.
//
// Сам выберет нужный тип, в зависимости от конфигурации сервера:
//  0. PsqlStorage
//  1. FileStorage
//  2. MemoryStorage
func NewStorager(cfg config.Config) (Storager, error) {
	switch getStoragerType(cfg) {
	case PsqlStorage:
		return postgres.NewPsqlStorage(cfg.DatabaseDSN)
	case FileStorage:
		file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDWR|os.O_CREATE, 0o777)
		if err != nil {
			return nil, err
		}
		return harddrive.NewFileStorage(file)
	default:
		return memory.NewMemoryStorage()
	}
}

func getStoragerType(cfg config.Config) StoragerType {
	if cfg.DatabaseDSN != "" {
		return PsqlStorage
	}
	if cfg.FileStoragePath != "" {
		return FileStorage
	}
	return MemoryStorage
}
