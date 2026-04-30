package repository

import (
	"context"
	"errors"
	"microblog/services/api/internal/models"

	"github.com/jackc/pgx/v5"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
	builder sq.StatementBuilderType
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query, args, err := r.builder.
		Insert("users").
		Columns("username", "email").
		Values(user.Username, user.Email).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, query, args...).Scan(&user.ID)
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*models.User, error) {
	user := &models.User{}
	query, args, err := r.builder.
		Select("id", "username", "email").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	query, args, err := r.builder.
		Select("id", "username", "email").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
