package repository

import (
	"EFms1Redis/pkg/models"
	"fmt"

	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type GetBookRepo struct {
	myArr  []models.Book
	Client redis.Client
}

func NewGetBookRepo(client *redis.Client) *GetBookRepo {
	return &GetBookRepo{myArr: make([]models.Book, 0), Client: *client}
}

func (r *GetBookRepo) GetBook(ctx context.Context, bookName string) (models.Book, error) {
	for i := 0; i < len(r.myArr); i++ {
		if r.myArr[i].BookName == bookName {
			return r.myArr[i], nil
		}
	}
	return models.Book{}, fmt.Errorf("element not found")
}

// ConsumeBook read user from the "example" stream and log it
func (r *GetBookRepo) ConsumeBook(stream string) {
	go func() {
		for {
			var err error
			var data []redis.XMessage
			data, err = r.Client.XRangeN(context.Background(), stream, "-", "+", 1).Result()
			if err != nil {
				logrus.Error(err)
			}
			for _, element := range data {
				dataFromStream := []byte(element.Values["data"].(string))
				var book = &models.Book{}
				err := json.Unmarshal(dataFromStream, book)
				if err != nil {
					logrus.Error(err)
					continue
				}
				r.myArr = append(r.myArr, *book)
				r.Client.XDel(context.Background(), stream, element.ID)
			}
		}
	}()
}
