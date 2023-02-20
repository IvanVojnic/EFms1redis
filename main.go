package main

import (
	"EFms1Redis/pkg/models"
	"EFms1Redis/pkg/repository"
	service2 "EFms1Redis/pkg/service"
	"encoding/json"

	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type GetBooks interface {
	GetBook(ctx context.Context, bookName string) (models.Book, error)
	GetBookFromMs(ctx context.Context, name string) (models.Book, error)
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
	kafkaConn := repository.NewKafkaConn(conn)
	go consumer(kafkaConn)

	connRMQ, err := amqp.Dial("amqp://rabbit:rabbit@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connRMQ.Close()

	chRMQ, err := connRMQ.Channel()
	failOnError(err, "Failed to open a channel")
	defer chRMQ.Close()

	q, err := chRMQ.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	msgs, err := chRMQ.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			var book models.Book
			err := json.Unmarshal(d.Body, &book)
			if err != nil {
				log.Panicf("error while getting a book, %s", err)
			}
			log.Printf("Received a message: %s, %v, %v", book.BookName, book.BookNew, book.BookYear)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	service := service2.NewBookSrv(rds, kafkaConn)
	handler := NewHandler(service)
	e.GET("/getBook", handler.getBook)
	e.Logger.Fatal(e.Start(":30000"))
}

func consumer(con *repository.KafkaConn) {
	err := con.CreateBookKafka()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error save book by kafka": err,
		}).Info("save book request")
	}
}

func (h *Handler) getBook(c echo.Context) error {
	bookName := c.QueryParam("name")
	book, err := h.serviceBook.GetBookFromMs(c.Request().Context(), bookName)
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

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
