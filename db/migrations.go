package db

import (
	"api-service/model"
	"context"
	"log"

	"github.com/uptrace/bun"
)

// RunMigrations выполняет создание всех необходимых таблиц
func RunMigrations(db *bun.DB) {
	ctx := context.Background()

	// Создаем таблицу пользователей
	if _, err := db.NewCreateTable().
		Model((*model.User)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}

	// Создаем таблицу постов
	if _, err := db.NewCreateTable().
		Model((*model.Post)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы постов: %v", err)
	}

	log.Println("Миграции завершены: все таблицы созданы или уже существуют.")
}
