package usecase

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity/converter"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/repository"
)

type PurchaseCase interface {
	GetMerchantNearby(ctx context.Context, params *converter.MerchanNearbyParams) (*[]converter.MerchanNearby, error)
}

type purchaseCase struct {
	pool  *pgxpool.Pool
	prepo *repository.PurchaseRepo
}

func NewPurchaseCase(pool *pgxpool.Pool) PurchaseCase {
	return &purchaseCase{
		pool:  pool,
		prepo: &repository.PurchaseRepo{},
	}
}

func (p *purchaseCase) GetMerchantNearby(ctx context.Context, params *converter.MerchanNearbyParams) (*[]converter.MerchanNearby, error) {
	coordinate := strings.Split(params.Coordinate, ",")
	if len(coordinate) != 2 {
		return nil, exception.BadRequest("Coordinate not valid")
	}

	lat, err := strconv.ParseFloat(coordinate[0], 64)
	if err != nil {
		return nil, exception.BadRequest("Coordinate not valid")
	}

	long, err := strconv.ParseFloat(coordinate[1], 64)
	if err != nil {
		return nil, exception.BadRequest("Coordinate not valid")
	}

	if params.Limit == 0 {
		params.Limit = 5
	}

	data := p.prepo.GetMerchantNearby(ctx, p.pool, lat, long, params)

	return &data, nil
}
