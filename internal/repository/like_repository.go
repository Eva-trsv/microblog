package repository

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

type LikeRepository struct {
	db *pgxpool.Pool
}

func (r *LikeRepository) DB() *pgxpool.Pool {
	return r.db
}

func NewLikeRepository(db *pgxpool.Pool) *LikeRepository {
	return &LikeRepository{db: db}
}

func (r *LikeRepository) AddLike(ctx context.Context, tx pgx.Tx, userID, postID int) error {

	insertQuery, insertArgs, err := sq.
		Insert("likes").
		Columns("user_id", "post_id").
		Values(userID, postID).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, insertQuery, insertArgs...)
	if err != nil {
		return err
	}

	updateQuery, updateArgs, err := sq.
		Update("posts").
		Set("like_count", sq.Expr("like_count + 1")).
		Where(sq.Eq{"id": postID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, updateQuery, updateArgs...)
	if err != nil {
		return err
	}

	return nil
}

func (r *LikeRepository) CountLikes(ctx context.Context, tx pgx.Tx, userID, postID int) (int, error) {
	var count int
	query, args, err := sq.
		Select("COUNT(*)").
		From("likes").
		Where(sq.Eq{"post_id": postID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}