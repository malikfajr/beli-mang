package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/pkg/password"
	"github.com/malikfajr/beli-mang/internal/repository"
)

type adminAuth struct {
	pool *pgxpool.Pool
}

func NewAdminAuth(pool *pgxpool.Pool) *adminAuth {
	return &adminAuth{
		pool: pool,
	}
}

func (a *adminAuth) Insert(ctx context.Context, payload *entity.User) error {
	payload.Password = password.Hash(payload.Password)

	tx, err := a.pool.Begin(ctx)
	if err != nil {
		panic(err)
	}
	defer tx.Rollback(ctx)

	adminRepo := &repository.AdminRepo{}
	if exist := adminRepo.EmailExistTx(ctx, tx, payload.Email); exist == true {
		return exception.Conflict("Email is exists")
	}

	if err := adminRepo.InsertTx(ctx, tx, payload); err != nil {
		return exception.Conflict("Username is exists")
	}

	tx.Commit(ctx)
	return nil
}

func (a *adminAuth) Login(ctx context.Context, payload *entity.UserLogin) (*entity.User, error) {
	adminRepo := &repository.AdminRepo{}

	user, err := adminRepo.GetByUsername(ctx, a.pool, payload.Username)
	if err != nil {
		return nil, exception.BadRequest("request doesn’t pass validation / password is wrong")
	}

	if password.Compare(user.Password, payload.Password) == false {
		return nil, exception.BadRequest("request doesn’t pass validation / password is wrong")
	}

	return user, nil
}
