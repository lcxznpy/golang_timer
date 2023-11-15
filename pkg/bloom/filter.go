package bloom

import (
	"context"
	"math"
	"xtimer/pkg/hash"
	"xtimer/pkg/redis"
)

// 布隆过滤器结构
type Filter struct {
	client     *redis.Client
	encryptor1 *hash.SHA1Encryptor
	encryptor2 *hash.Murmur3Encyptor
}

// 布隆过滤器函数
func NewFilter(client *redis.Client, e1 *hash.SHA1Encryptor, e2 *hash.Murmur3Encyptor) *Filter {
	return &Filter{
		client:     client,
		encryptor1: e1,
		encryptor2: e2,
	}
}

// 布隆过滤器用来判断 目标定时任务有没有被执行
func (f *Filter) Exist(ctx context.Context, key, val string) (bool, error) {
	//从redis bitmap找
	rawVal1 := f.encryptor1.Encrypt(val)
	//如果存在就返回
	if exist, err := f.client.GetBit(ctx, key, int32(rawVal1%math.MaxInt32)); exist || err != nil {
		return exist, err
	}
	rawVal2 := f.encryptor2.Encrypt(val)
	return f.client.GetBit(ctx, key, int32(rawVal2%math.MaxInt32))
}

// 定时任务执行后 在bitmap中存值
func (f *Filter) Set(ctx context.Context, key, val string, expireSeconds int64) error {
	// 判断key存不存在 ，如果不存在，就要设置过期时间，不通过事务保证原子性
	existed, _ := f.client.Exists(ctx, key)

	// 算出两个 hash函数的val，分别set
	rawVal1, rawVal2 := f.encryptor1.Encrypt(val), f.encryptor2.Encrypt(val)

	_, err := f.client.Transaction(ctx, redis.NewSetBitCommand(key, int32(rawVal1%math.MaxInt32), 1),
		redis.NewSetBitCommand(key, int32(rawVal2%math.MaxInt32), 1))

	if !existed {
		_ = f.client.Expire(ctx, key, expireSeconds)
	}
	return err
}
