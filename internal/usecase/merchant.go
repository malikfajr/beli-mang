package usecase

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/repository"
	"github.com/oklog/ulid/v2"
)

type manageMerchant struct {
	pool *pgxpool.Pool
}

func NewManageMerchant(pool *pgxpool.Pool) *manageMerchant {
	return &manageMerchant{
		pool: pool,
	}
}

func (m *manageMerchant) Create(ctx context.Context, username string, payload *entity.AddMerchantPayload) (*entity.Merchant, error) {
	id := ulid.Make()

	merchant := &entity.Merchant{
		Id:       id.String(),
		Username: username,
		Category: payload.Category,
		ImageUrl: payload.ImageUrl,
		Name:     payload.Name,
		Location: payload.Location,
	}

	merchantRepo := &repository.MerchantRepo{}
	err := merchantRepo.Insert(ctx, m.pool, merchant)
	if err != nil {
		return nil, exception.ServerError(err.Error())
	}

	return merchant, nil
}

func (m *manageMerchant) GetAll(ctx context.Context, username string, params *entity.MerchantParams) (*[]entity.Merchant, error) {
	if params.Limit == 0 {
		params.Limit = 5
	}

	if params.CreatedAt != "asc" || params.CreatedAt != "desc" {
		params.CreatedAt = ""
	}

	merchantRepo := &repository.MerchantRepo{}
	merchants := merchantRepo.GetAll(ctx, m.pool, username, params)

	return &merchants, nil
}
