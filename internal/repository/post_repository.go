package repository

import (
	"context"
	"microblog/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	db *pgxpool.Pool
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query, args, err := sq.
		Insert("posts").
		Columns("author_id", "content", "like_count").
		Values(post.AuthorID, post.Content, post.LikeCount).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, query, args...).Scan(&post.ID)
}

func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID int) ([]*models.Post, error) {
	query, args, err := sq.
		Select("id", "author_id", "content", "like_count").
		From("posts").
		Where(sq.Eq{"author_id": authorID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*models.Post
	for rows.Next() {
		post := &models.Post{}
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Content, &post.LikeCount); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func (r *PostRepository) GetPostByID(ctx context.Context, postID int) (*models.Post, error) {
	query, args, err := sq.
		Select("id", "author_id", "content", "like_count").
		From("posts").
		Where(sq.Eq{"id": postID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, err
	}

	post := &models.Post{}
	err = r.db.QueryRow(ctx, query, args...).Scan(&post.ID, &post.AuthorID, &post.Content, &post.LikeCount)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *PostRepository) Delete(ctx context.Context, postID int) error {
	query, args, err := sq.
		Delete("posts").
		Where(sq.Eq{"id": postID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}