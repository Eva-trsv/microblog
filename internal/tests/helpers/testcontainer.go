package helpers

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer представляет тестовый контейнер PostgreSQL
type PostgresContainer struct {
	Container *postgres.PostgresContainer
	Pool      *pgxpool.Pool
	DSN       string
}

// SetupPostgresContainer создаёт и запускает контейнер PostgreSQL для тестов
func SetupPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	// Находим путь к миграциям (относительно текущего файла)
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	migrationsPath := filepath.Join(basepath, "..", "..", "..", "migrations")

	// Создаём контейнер PostgreSQL
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(filepath.Join(migrationsPath, "20260224143913_init.sql"),
			filepath.Join(migrationsPath, "20260224160000_add_author_id_and_likes.sql"),
			filepath.Join(migrationsPath, "20260301190000_add_indexes.sql"),
			filepath.Join(migrationsPath, "20260314120000_partitioning_to_posts.sql.sql")),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Получаем строку подключения
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Подключаемся к базе
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &PostgresContainer{
		Container: pgContainer,
		Pool:      pool,
		DSN:       connStr,
	}, nil
}

// Cleanup закрывает соединение и останавливает контейнер
func (pc *PostgresContainer) Cleanup(ctx context.Context) error {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		return pc.Container.Terminate(ctx)
	}
	return nil
}

// TruncateTables очищает все таблицы между тестами
func (pc *PostgresContainer) TruncateTables(ctx context.Context) error {
	tables := []string{"likes", "posts", "users"}

	for _, table := range tables {
		_, err := pc.Pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}
	return nil
}
