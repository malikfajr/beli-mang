package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
)

type AdminRepo struct{}

func (r *AdminRepo) EmailExistTx(ctx context.Context, tx pgx.Tx, email string) bool {
	var exist int
	query := "SELECT 1 FROM users WHERE email = $1 AND admin = true LIMIT 1;"

	err := tx.QueryRow(ctx, query, email).Scan(&exist)
	if err != nil {
		return false
	}

	return true
}

func (r *AdminRepo) InsertTx(ctx context.Context, tx pgx.Tx, user *entity.User) error {
	query := `INSERT INTO users(admin, username, password, email) VALUES(true, @username, @password, @email) ON CONFLICT DO NOTHING`
	args := pgx.NamedArgs{
		"username": user.Username,
		"password": user.Password,
		"email":    user.Email,
	}

	tag, err := tx.Exec(ctx, query, args)
	if err != nil {
		panic(err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("Username already exists")
	}

	return nil
}

func (r *AdminRepo) GetByUsername(ctx context.Context, pool *pgxpool.Pool, username string) (*entity.User, error) {
	var user = &entity.User{IsAdmin: true}
	query := "SELECT username, password, email FROM users WHERE username = $1 AND admin = true LIMIT 1;"

	err := pool.QueryRow(ctx, query, username).Scan(&user.Username, &user.Password, &user.Email)
	if err != nil {
		return nil, errors.New("Account not found!")
	}

	return user, nil
}
