package redis

import (
	"context"
	"errors"
	"xtimer/common/utils"

	"github.com/gomodule/redigo/redis"
)

const ftimerLockKeyPrefix = "FTIMER_LOCK_PREFIX_"

type DistributeLocker interface {
	Lock(context.Context, int64) error
	Unlock(context.Context) error
	ExpireLock(ctx context.Context, expireSeconds int64) error
}

// ReentrantDistributeLock 可重入分布式锁.
type ReentrantDistributeLock struct {
	key    string
	token  string
	client *Client
}

// 可重入分布式锁 构造函数
func NewReentrantDistributeLock(key string, client *Client) *ReentrantDistributeLock {
	return &ReentrantDistributeLock{
		key:    key,
		token:  utils.GetProcessAndGoroutineIDStr(),
		client: client,
	}

}

// 给分布式锁加锁
func (r *ReentrantDistributeLock) Lock(ctx context.Context, expireSeconds int64) error {
	// 查看锁是否是自己的
	res, err := r.client.Get(ctx, r.key)
	if err != nil && !errors.Is(err, redis.ErrNil) {
		return err
	}
	//是自己的，不用再加锁了
	if res == r.token {
		return nil
	}
	//锁不是自己的，但不清楚是不是别人的,那我先尝试set值，如果成功就获得锁，反之锁是别人的，失败
	reply, err := r.client.SetNX(ctx, r.key, r.token, expireSeconds)
	if err != nil {
		return err
	}
	re, _ := reply.(int64)
	if re != 1 {
		return errors.New("locker is acquiered by others")
	}
	return nil

}

// 解锁,通过lua脚本保证事务一致性
func (r *ReentrantDistributeLock) Unlock(ctx context.Context) error {
	keysAndArgs := []interface{}{r.getLockKey(), r.token}
	//查询这把锁是不是你的，是的话就删除key-value对
	reply, err := r.client.Eval(ctx, LuaCheckAndDeleteDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}
	if ret, _ := reply.(int64); ret != 1 {
		//这把锁不是你的，解锁不了
		return errors.New("can not unlock without ownership of locker")
	}
	return nil
}

// 延长锁的过期时间
func (r *ReentrantDistributeLock) ExpireLock(ctx context.Context, expireSeconds int64) error {
	keysAndArgs := []interface{}{r.getLockKey(), r.token, expireSeconds}
	//查询这把锁是不是你的，是的话就删除key-value对
	reply, err := r.client.Eval(ctx, LuaCheckAndExpireDistributionLock, 1, keysAndArgs)
	if err != nil {
		return err
	}
	if ret, _ := reply.(int64); ret != 1 {
		//这把锁不是你的，解锁不了
		return errors.New("can not expire without ownership of locker")
	}
	return nil
}

func (r *ReentrantDistributeLock) getLockKey() string {
	return ftimerLockKeyPrefix + r.key
}
