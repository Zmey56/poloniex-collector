package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer содержит контейнер и пул подключений к PostgreSQL
type PostgresContainer struct {
	Container testcontainers.Container
	Pool      *pgxpool.Pool
}

// NewPostgresContainer создает новый тестовый контейнер PostgreSQL
func NewPostgresContainer(t *testing.T) (*PostgresContainer, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:14-alpine",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("5432/tcp"),
			wait.ForLog("database system is ready to accept connections"),
		).WithStartupTimeout(30 * time.Second),
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
		},
		AutoRemove: false, // Отключаем автоматическое удаление контейнера
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// Проверяем, не завершился ли контейнер
	state, err := container.State(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container state: %v", err)
	}
	if !state.Running {
		container.Terminate(ctx)
		return nil, fmt.Errorf("container exited unexpectedly")
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get host: %v", err)
	}

	connString := fmt.Sprintf(
		"postgres://test:test@%s:%s/testdb?sslmode=disable",
		host,
		mappedPort.Port(),
	)
	fmt.Printf("Connection string: %s\n", connString)

	// Добавляем задержку перед подключением
	time.Sleep(2 * time.Second)

	// Создаем пул подключений с повторными попытками
	var pool *pgxpool.Pool
	for i := 0; i < 5; i++ {
		pool, err = pgxpool.Connect(ctx, connString)
		if err == nil {
			break
		}
		fmt.Printf("Retrying DB connection... (%d/5)\n", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	// Применяем миграции
	if err := applyMigrations(ctx, pool); err != nil {
		pool.Close()
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to apply migrations: %v", err)
	}

	return &PostgresContainer{
		Container: container,
		Pool:      pool,
	}, nil
}

// Close закрывает пул подключений и останавливает контейнер
func (pc *PostgresContainer) Close() {
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Container != nil {
		pc.Container.Terminate(context.Background())
	}
}

// applyMigrations применяет миграции к тестовой базе данных
func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS klines (
            id SERIAL PRIMARY KEY,
            pair VARCHAR(20) NOT NULL,
            interval VARCHAR(10) NOT NULL,
            open DECIMAL(20, 8) NOT NULL,
            high DECIMAL(20, 8) NOT NULL,
            low DECIMAL(20, 8) NOT NULL,
            close DECIMAL(20, 8) NOT NULL,
            utc_begin BIGINT NOT NULL,
            utc_end BIGINT NOT NULL,
            volume_bs JSONB NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(pair, interval, utc_begin)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_klines_pair_interval_utc ON klines(pair, interval, utc_begin)`,

		`CREATE TABLE IF NOT EXISTS trades (
            id BIGSERIAL PRIMARY KEY,
            tid VARCHAR(255) NOT NULL,
            pair VARCHAR(20) NOT NULL,
            price DECIMAL(20, 8) NOT NULL,
            amount DECIMAL(20, 8) NOT NULL,
            quantity DECIMAL(20, 8) NOT NULL,
            side VARCHAR(4) NOT NULL,
            timestamp BIGINT NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(tid, pair)
        )`,
		`CREATE INDEX IF NOT EXISTS idx_trades_pair_timestamp ON trades(pair, timestamp)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration: %v", err)
		}
	}

	return nil
}
