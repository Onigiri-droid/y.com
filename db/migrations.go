package db

import (
	"api-service/model"
	"context"
	"log"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/extra/bundebug"
)

// RunMigrations выполняет создание и обновление всех необходимых таблиц
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

	// Обновляем таблицу пользователей: добавляем недостающие столбцы
	if _, err := db.ExecContext(ctx, `
		ALTER TABLE users
		ADD COLUMN IF NOT EXISTS avatar VARCHAR(255),
		ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'offline',
		ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ADD COLUMN IF NOT EXISTS last_seen TIMESTAMP,
		ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user',
		ADD COLUMN IF NOT EXISTS bio TEXT;
	`); err != nil {
		log.Fatalf("Ошибка обновления таблицы пользователей: %v", err)
	}
	log.Println("Обновлена таблица пользователей: добавлены avatar, status, created_at, last_seen, role, bio.")

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

	// Обновляем внешний ключ для каскадного удаления постов
	if _, err := db.ExecContext(ctx, `
    -- Удаляем записи из posts, где user_id ссылается на несуществующих пользователей
    DELETE FROM posts
    WHERE user_id NOT IN (SELECT id FROM users);
    
    -- Обновляем внешний ключ с каскадным удалением
    ALTER TABLE posts
    DROP CONSTRAINT IF EXISTS posts_user_id_fkey,
    ADD CONSTRAINT posts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
	`); err != nil {
    log.Fatalf("Ошибка обновления внешнего ключа для каскадного удаления: %v", err)
	}
	log.Println("Внешний ключ таблицы posts обновлен для каскадного удаления.")

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
