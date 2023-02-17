// Package service for book
package service

import (
	"EFms1Redis/pkg/models"
	"context"
)

type GetBooks interface {
	GetBook(ctx context.Context, name string) (models.Book, error)
}

type BookFromMS interface {
	GetSavedBook(ctx context.Context, name string) (models.Book, error)
}

type GetBookSrv struct {
	repo   GetBooks
	kafkaR BookFromMS
}

func NewBookSrv(repo GetBooks, kafkaR BookFromMS) *GetBookSrv {
	return &GetBookSrv{repo: repo, kafkaR: kafkaR}
}

func (s *GetBookSrv) GetBook(ctx context.Context, bookName string) (models.Book, error) {
	return s.repo.GetBook(ctx, bookName)
}

func (s *GetBookSrv) GetBookFromMs(ctx context.Context, name string) (models.Book, error) {
	return s.kafkaR.GetSavedBook(ctx, name)
}
