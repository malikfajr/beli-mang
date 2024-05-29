package usecase

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/malikfajr/beli-mang/internal/entity"
	"github.com/malikfajr/beli-mang/internal/exception"
	"github.com/malikfajr/beli-mang/internal/repository"
	"github.com/oklog/ulid/v2"
)

type manageMerchant struct {
	pool *pgxpool.Pool
	id   map[string]string
	sync.Mutex
}

type ManageMerchant interface {
	Create(ctx context.Context, username string, payload *entity.AddMerchantPayload) (*entity.Merchant, error)
	GetAll(ctx context.Context, username string, params *entity.MerchantParams) (*[]entity.Merchant, error)
	AddProduct(ctx context.Context, merchantId string, payload *entity.AddProductPayload) (*entity.Product, error)
	GetProducts(ctx context.Context, username string, params *entity.ProductParams) (*[]entity.Product, error)
	ResetData()
}

func NewManageMerchant(pool *pgxpool.Pool) ManageMerchant {
	return &manageMerchant{
		pool: pool,
		id:   map[string]string{},
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

	m.Lock()
	defer m.Unlock()
	m.id[merchant.Id] = username

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

func (m *manageMerchant) AddProduct(ctx context.Context, merchantId string, payload *entity.AddProductPayload) (*entity.Product, error) {
	if err := m.isFound(nil, merchantId); err != nil {
		return nil, err
	}

	product := &entity.Product{
		Id:         ulid.Make().String(),
		MerchantId: merchantId,
		Name:       payload.Name,
		Category:   payload.Category,
		Price:      payload.Price,
		ImageUrl:   payload.ImageUrl,
	}

	merchantRepo := &repository.MerchantRepo{}
	err := merchantRepo.AddProduct(ctx, m.pool, product)
	if err != nil {
		panic(err)
	}

	return product, nil
}

func (m *manageMerchant) GetProducts(ctx context.Context, username string, params *entity.ProductParams) (*[]entity.Product, error) {
	if err := m.isFound(&username, params.MerchantId); err != nil {
		return nil, err
	}

	if params.Limit <= 0 {
		params.Limit = 5
	}

	if params.CreatedAt != "asc" || params.CreatedAt != "desc" {
		params.CreatedAt = ""
	}

	merchantRepo := &repository.MerchantRepo{}
	products := merchantRepo.GetProducts(ctx, m.pool, params)

	return &products, nil
}

func (m *manageMerchant) isFound(username *string, merchantId string) error {
	_, err := ulid.Parse(merchantId)
	if err != nil {
		return exception.NotFound("merchantId not found")
	}

	m.Lock()
	defer m.Unlock()

	if user, ok := m.id[merchantId]; username != nil && ok {
		if user != *username {
			return exception.NotFound("merchantId not found")
		}

		return nil
	}

	merchanRepo := &repository.MerchantRepo{}
	merchant, err := merchanRepo.GetById(context.Background(), m.pool, merchantId)
	if err != nil {
		return exception.NotFound("merchantId not found")
	}

	m.id[merchant.Id] = merchant.Username

	if username != nil && *username != merchant.Username {
		return exception.NotFound("merchantId not found")
	}

	return nil
}

func (m *manageMerchant) ResetData() {
	m.Lock()
	defer m.Unlock()

	m.id = make(map[string]string, 0)
}
