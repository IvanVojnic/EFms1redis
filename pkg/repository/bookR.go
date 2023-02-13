package repository

import (
	"EFms1Redis/pkg/models"

	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type GetBookRepo struct {
	myArr []models.Book
}

func NewGetBookRepo() *GetBookRepo {
	return &GetBookRepo{myArr: make([]models.Book, 0)}
}

func (r *GetBookRepo) GetBook(ctx context.Context, bookName string) (models.Book, error) {
	return r.myArr[0], nil
}

// ConsumeUser read user from the "example" stream and log it
func (r *GetBookRepo) ConsumeUser(c context.Context, stream string) {
	go func() {
		for {
			var err error
			var data []redis.XMessage
			data, err = c.Client.XRangeN(context.Background(), stream, "-", "+", 1).Result()
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
				c.Client.XDel(context.Background(), stream, element.ID)
			}
		}
	}()
}
