package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db      *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:      db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// 🔹 1. Проверка — обработано ли событие
func (r *Repository) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	query, args, err := r.builder.
		Select("1").
		From("processed_events").
		Where(sq.Eq{"event_id": eventID}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, err
	}

	var dummy int
	err = r.db.QueryRow(ctx, query, args...).Scan(&dummy)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// 🔹 2. Сохранить event_id
func (r *Repository) SaveProcessed(ctx context.Context, eventID string) error {
	query, args, err := r.builder.
		Insert("processed_events").
		Columns("event_id").
		Values(eventID).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

//
// 🔹 3. Увеличить лайки (UPSERT)
//
func (r *Repository) IncrementLike(ctx context.Context, postID int) error {
	query, args, err := r.builder.
		Insert("post_like_counters").
		Columns("post_id", "like_count").
		Values(postID, 1).
		Suffix("ON CONFLICT (post_id) DO UPDATE SET like_count = post_like_counters.like_count + 1").
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}

//
// 🔹 4. Получить лайки
//
func (r *Repository) GetLikes(ctx context.Context, postID int) (int, error) {
	var count int

	query, args, err := r.builder.
		Select("like_count").
		From("post_like_counters").
		Where(sq.Eq{"post_id": postID}).
		ToSql()
	if err != nil {
		return 0, err
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}