package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/password"
	"github.com/malikfajr/beli-mang/internal/repository"
)

type userAuth struct {
	pool *pgxpool.Pool
}

func NewUserAuth(pool *pgxpool.Pool) *userAuth {
	return &userAuth{
		pool: pool,
	}
}

func (a *userAuth) Insert(ctx context.Context, payload *entity.User) error {
	payload.Password = password.Hash(payload.Password)

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)

	userRepo := &repository.UserRepo{}
	if exist := userRepo.EmailExistTx(ctx, tx, payload.Email); exist == true {
		return exception.Conflict("Email is exists")
	}

	if err := userRepo.InsertTx(ctx, tx, payload); err != nil {
		return exception.Conflict("Username is exists")
	}

	tx.Commit(ctx)
	return nil
}

func (a *userAuth) Login(ctx context.Context, payload *entity.UserLogin) (*entity.User, error) {
	userRepo := &repository.UserRepo{}

	user, err := userRepo.GetByUsername(ctx, a.pool, payload.Username)
	if err != nil {
		return nil, exception.BadRequest("request doesn’t pass validation / password is wrong")
	}

	if password.Compare(user.Password, payload.Password) == false {
		return nil, exception.BadRequest("request doesn’t pass validation / password is wrong")
	}

	return user, nil
}
