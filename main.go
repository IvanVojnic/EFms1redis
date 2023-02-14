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
	"github.com/sirupsen/logrus"
)

type GetBooks interface {
	GetBook(ctx context.Context, bookName string) (models.Book, error)
}

type Handler struct {
	serviceBook GetBooks
}

func main() {
	e := echo.New()

	logger := logrus.New()
	logger.Out = os.Stdout
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			logrus.WithFields(logrus.Fields{
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

	st := rdb.Ping(context.Background())
	if st.Err() != nil {
		logrus.Fatal()
	}

	defer rdb.Close()
	rds := repository.NewGetBookRepo(rdb)
	rds.ConsumeUser("example")
	service := service2.NewBookSrv(rds)
	handler := NewHandler(service)
	e.GET("/getBook", handler.getBook)
	e.Logger.Fatal(e.Start(":30000"))
}

func (h *Handler) getBook(c echo.Context) error {
	bookName := c.QueryParam("name")
	book, err := h.serviceBook.GetBook(c.Request().Context(), bookName)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error get book": err,
			"book":           book,
		}).Info("GET BOOK request")
		return echo.NewHTTPError(http.StatusBadRequest, "cannot get book")
	}
	return c.JSON(http.StatusOK, book)
}

func NewHandler(serviceBook GetBooks) *Handler {
	return &Handler{
		serviceBook: serviceBook,
	}
}
