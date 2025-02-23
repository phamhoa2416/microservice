package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"microservices/model"
)

type RedisRepository struct {
	Client *redis.Client
}

func (r *RedisRepository) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	key := orderIDKey(order.OrderID)

	res := r.Client.SetNX(ctx, key, string(data), 0)
	if res.Err() != nil {
		return fmt.Errorf("failed to insert order: %v", res.Err())
	}

	return nil
}

func (r *RedisRepository) FindById(ctx context.Context, orderId uint64) (model.Order, error) {
	key := orderIDKey(orderId)

	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, NotExistErr
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to fetch order: %v", err)
	}

	var order model.Order
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to unmarshal order: %v", err)
	}

	return order, nil
}

func (r *RedisRepository) Delete(ctx context.Context, orderId uint64) error {
	key := orderIDKey(orderId)

	err := r.Client.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		return NotExistErr
	} else if err != nil {
		return fmt.Errorf("failed to delete order: %v", err)
	}

	return nil
}

func (r *RedisRepository) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %v", err)
	}

	key := orderIDKey(order.OrderID)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return NotExistErr
	} else if err != nil {
		return fmt.Errorf("failed to update order: %v", err)
	}

	return nil
}

func (r *RedisRepository) FindAll(ct context.Context) ([]model.Order, error) {
	return nil, nil
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

var NotExistErr = errors.New("order not exist")

type FindAllPage struct {
	Size   uint
	Offset uint
}
