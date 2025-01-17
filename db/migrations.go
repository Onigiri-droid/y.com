package db

import (
	"api-service/model"
	"context"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
)

// RunMigrations выполняет создание всех необходимых таблиц и триггеров
func RunMigrations(db *bun.DB) {
	// Включение отладки SQL-запросов
	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))

	ctx := context.Background()

	log.Println("Начало миграций...")

	// Создаем таблицу пользователей
	if _, err := db.NewCreateTable().
		Model((*model.User)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы пользователей: %v", err)
	}
	log.Println("Таблица пользователей создана.")

	// Создаем таблицу постов
	if _, err := db.NewCreateTable().
		Model((*model.Post)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы постов: %v", err)
	}
	log.Println("Таблица постов создана.")

	// Обновляем таблицу постов
	if _, err := db.ExecContext(ctx, `
		ALTER TABLE posts
		ADD COLUMN IF NOT EXISTS likes_count INT DEFAULT 0,
		ADD COLUMN IF NOT EXISTS comments_count INT DEFAULT 0,
		ADD COLUMN IF NOT EXISTS reposts_count INT DEFAULT 0;
	`); err != nil {
		log.Fatalf("Ошибка обновления таблицы постов: %v", err)
	}
	log.Println("Обновлена таблица постов: добавлены likes_count, comments_count, reposts_count.")

	// Создаем таблицу тегов
	if _, err := db.NewCreateTable().
		Model((*model.Tag)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		log.Fatalf("Ошибка создания таблицы тегов: %v", err)
	}
	log.Println("Таблица тегов создана.")

	// Создаем таблицы для связанных моделей
	tables := []interface{}{
		(*model.PostLike)(nil),
		(*model.Comment)(nil),
		(*model.PostTag)(nil),
		(*model.Media)(nil),
		(*model.Repost)(nil),
	}

	for _, table := range tables {
		if _, err := db.NewCreateTable().
			Model(table).
			IfNotExists().
			Exec(ctx); err != nil {
			log.Fatalf("Ошибка создания таблицы %T: %v", table, err)
		}
		log.Printf("Таблица %T создана.", table)
	}

	log.Println("Все таблицы созданы или уже существуют.")
}
