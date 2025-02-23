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
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	key := orderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to insert order: %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to order set: %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

func (r *RedisRepository) FindById(ctx context.Context, orderId uint64) (model.Order, error) {
	key := orderIDKey(orderId)

	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return model.Order{}, NotExistErr
	} else if err != nil {
		return model.Order{}, fmt.Errorf("failed to fetch order: %w", err)
	}

	var order model.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return model.Order{}, fmt.Errorf("failed to unmarshal order: %w", err)
	}

	return order, nil
}

func (r *RedisRepository) DeleteById(ctx context.Context, orderId uint64) error {
	key := orderIDKey(orderId)
	txn := r.Client.TxPipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return NotExistErr
	} else if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	if _, err := txn.SRem(ctx, "orders", key).Result(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from order sets: %v", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %v", err)
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

func (r *RedisRepository) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders id: %v", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %v", err)
	}

	orders := []model.Order{}

	for _, x := range xs {
		if x == nil {
			continue
		}

		x, ok := x.(string)
		if !ok {
			return FindResult{}, fmt.Errorf("failed to unmarshal order: %v", err)
		}

		var order model.Order
		if err := json.Unmarshal([]byte(x), &order); err != nil {
			return FindResult{}, fmt.Errorf("failed to unmarshal order: %v", err)
		}

		orders = append(orders, order)
	}

	return FindResult{
		Orders: orders,
		Cursor: cursor,
	}, nil
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}

var NotExistErr = errors.New("order not exist")

type FindAllPage struct {
	Size   uint64
	Offset uint64
}

type FindResult struct {
	Orders []model.Order
	Cursor uint64
}
