package main

import (
	"EFms1Redis/pkg/models"
	"EFms1Redis/pkg/repository"
	service2 "EFms1Redis/pkg/service"
	"context"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

type GetBooks interface {
	GetBook(ctx context.Context, bookName string) (models.Book, error)
}

type Handler struct {
	serviceBook GetBooks
}

func main() {
	e := echo.New()

	logger := log.New()
	logger.Out = os.Stdout
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			log.WithFields(log.Fields{
				"URI":    values.URI,
				"status": values.Status,
			}).Info("request")
			return nil
		},
	}))
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
	})
	defer rdb.Close()
	rds := repository.NewGetBookRepo()
	service := service2.NewBookSrv(rds)
	handler := NewHandler(service)
	e.POST("/getBook", handler.getBook)
	e.Logger.Fatal(e.Start(":30000"))
}

func (h *Handler) getBook(c echo.Context) error {
	return echo.NewHTTPError(http.StatusBadRequest, "failed...")
}

func NewHandler(serviceBook GetBooks) *Handler {
	return &Handler{
		serviceBook: serviceBook,
	}
}
