package repository

import (
	"EFms1Redis/pkg/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"sync"
)

type GetBookRepo struct {
	mut   sync.Mutex
	myArr []models.Book
}

func (r *GetBookRepo) GetBook(ctx context.Context, bookName string) (models.Book, error) {
	/*myCache := cache.New(&cache.Options{
		Redis: r.Client,
	})
	book := &models.Book{}
	err := myCache.Get(ctx, bookName, book)
	if err != nil {
		return *book, fmt.Errorf("redis - GetByLogin - Get: %w", err)
	}*/
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
				mycache := cache.New(&cache.Options{
					Redis: r.Client,
				})

				err := mycache.Set(&cache.Item{
					Ctx:   ctx,
					Key:   book.BookName,
					Value: book,
				})
				if err != nil {
					return fmt.Errorf("redis - CreateBook - Set: %w", err)
				}
				return nil
				c.Client.XDel(context.Background(), stream, element.ID)
			}
		}
	}()
}
