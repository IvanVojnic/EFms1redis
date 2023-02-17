package repository

import (
	"EFms1Redis/pkg/models"
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaConn struct with kafka connection
type KafkaConn struct {
	readerConn *kafka.Conn
	writerConn *kafka.Conn
	booksArr   []models.Book
}

// NewKafkaConn used to init kafka connection
func NewKafkaConn(conn *kafka.Conn) *KafkaConn {
	return &KafkaConn{readerConn: conn, writerConn: conn, booksArr: make([]models.Book, 0)}
}

func (k *KafkaConn) GetSavedBook(ctx context.Context, bookName string) (models.Book, error) {
	for i := 0; i < len(k.booksArr); i++ {
		if k.booksArr[i].BookName == bookName {
			return k.booksArr[i], nil
		}
	}
	return models.Book{}, fmt.Errorf("element not found")
}

// CreateBookKafka used to send book to another ms to create book
func (k *KafkaConn) CreateBookKafka() error {
	batch := k.readerConn.ReadBatch(0, 1e6) // fetch 10KB min, 1MB max
	var book models.Book
	for {
		bookBytes, errRM := batch.ReadMessage()
		if errRM != nil {
			break
		}
		errUnMarsh := json.Unmarshal(bookBytes.Value, &book)
		if errUnMarsh != nil {
			break
		}
	}
	k.booksArr = append(k.booksArr, book)
	if err := batch.Close(); err != nil {
		return fmt.Errorf("error while getting book, %s", err)
	}
	return nil
}
