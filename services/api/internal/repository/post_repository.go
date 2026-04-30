package repository

import (
	"context"
	"microblog/services/api/internal/models"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	db *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewPostRepository(db *pgxpool.Pool) *PostRepository {
	return &PostRepository{
		db: db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PostRepository) Create(ctx context.Context, post *models.Post) error {
	query, args, err := r.builder.
		Insert("posts").
		Columns("author_id", "content", "like_count").
		Values(post.AuthorID, post.Content, post.LikeCount).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, query, args...).Scan(&post.ID)
}

func (r *PostRepository) GetByAuthorID(ctx context.Context, authorID int) ([]*models.Post, error) {
	query, args, err := r.builder.
		Select("id", "author_id", "content", "like_count").
		From("posts").
		Where(sq.Eq{"author_id": authorID}).
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
	query, args, err := r.builder.
		Select("id", "author_id", "content", "like_count").
		From("posts").
		Where(sq.Eq{"id": postID}).
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
	query, args, err := r.builder.
		Delete("posts").
		Where(sq.Eq{"id": postID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	return err
}