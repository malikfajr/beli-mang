package db

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getMaxPool(ctx context.Context, url string) (int, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return 0, err
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return 0, err
	}

	var getMaxConn string
	query := "SHOW max_connections;"

	if err := pool.QueryRow(ctx, query).Scan(&getMaxConn); err != nil {
		return 0, err
	}

	maxConn, err := strconv.ParseFloat(getMaxConn, 64)
	if err != nil {
		return 0, err
	}

	maxConnPool := int(math.Floor(maxConn * 0.9))
	if maxConn < 1 {
		return 1, nil
	}

	return maxConnPool, nil
}

func NewPool(ctx context.Context, url string) *pgxpool.Pool {
	pgConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Println("db url = " + url)
		log.Fatal("Cannot parsing database url", err)
		os.Exit(1)
	}

	maxConn, err := getMaxPool(ctx, url)
	if err != nil {
		log.Fatal("Cannot get max connection", err)
		os.Exit(1)
	}

	pgConfig.MaxConns = int32(maxConn)

	pool, err := pgxpool.NewWithConfig(ctx, pgConfig)
	if err != nil {
		log.Fatal("Cannot connect database.", err)
		os.Exit(1)
	}

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("Cannot connect database.", err)
		os.Exit(1)
	}

	return pool
}

func Address() string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?%s", os.Getenv("DB_USERNAME"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"), os.Getenv("DB_PARAMS"))
}

func Escape(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
