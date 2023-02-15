package main

import (
	"EFms1Redis/pkg/models"
	"EFms1Redis/pkg/repository"
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
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
	rds.ConsumeBook("example")

	ch := make(chan int, 1)
	consumer(ch)

	//service := service2.NewBookSrv(rds)
	//handler := NewHandler(service)
	//e.GET("/getBook", handler.getBook)
	//e.Logger.Fatal(e.Start(":30000"))
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

func consumer(ch chan int) {
	network := "tcp"
	address := "127.0.0.1:9092"
	topic := "myTopic"
	partition := 0
	conn, errKafka := kafka.DialLeader(context.Background(), network, address, topic, partition)
	if errKafka != nil {
		logrus.WithFields(logrus.Fields{
			"Error connection to database rep.NewPostgresDB()": errKafka,
		}).Fatal("DB ERROR CONNECTION")
	}

	kafkaReader := repository.NewKafkaConn(conn)
	err := kafkaReader.CreateBookKafka()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error save book by kafka": err,
		}).Info("save book request")
	}
}
