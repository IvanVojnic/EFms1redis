// Package service for book
package service

import (
	"EFms1Redis/pkg/models"
	"context"
)

type GetBooks interface {
	GetBook(ctx context.Context, name string) (models.Book, error)
}

type GetBookSrv struct {
	repo GetBooks
}

func NewBookSrv(repo GetBooks) *GetBookSrv {
	return &GetBookSrv{repo: repo}
}

func (s *GetBookSrv) GetBook(ctx context.Context, bookName string) (models.Book, error) {
	return models.Book{}, nil
}
