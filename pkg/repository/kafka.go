package repository

import (
	"fmt"

	"github.com/segmentio/kafka-go"
)

// KafkaConn struct with kafka connection
type KafkaConn struct {
	readerConn *kafka.Conn
}

// NewKafkaConn used to init kafka connection
func NewKafkaConn(conn *kafka.Conn) *KafkaConn {
	return &KafkaConn{readerConn: conn}
}

// CreateBookKafka used to send book to another ms to create book
func (k *KafkaConn) CreateBookKafka() error {
	batch := k.readerConn.ReadBatch(0, 1e6) // fetch 10KB min, 1MB max
	b := make([]byte, 10e3)                 // 10KB max per message
	for {
		n, err := batch.Read(b)
		if err != nil {
			break
		}
		fmt.Println(string(b[:n]))
	}
	if err := batch.Close(); err != nil {
		return fmt.Errorf("error while getting book, %s", err)
	}
	return nil
}
