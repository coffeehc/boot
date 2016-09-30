package redisservice

import (
	"fmt"
	"time"

	"github.com/coffeehc/microserviceboot/base"
	"gopkg.in/redis.v3"
)

type RedisService interface {
	Auth(password string) *redis.StatusCmd
	Echo(message string) *redis.StringCmd
	Ping() *redis.StatusCmd
	//Quit() *StatusCmd
	Select(index int64) *redis.StatusCmd
	Del(keys ...string) *redis.IntCmd
	Dump(key string) *redis.StringCmd
	Exists(key string) *redis.BoolCmd
	Expire(key string, expiration time.Duration) *redis.BoolCmd
	ExpireAt(key string, tm time.Time) *redis.BoolCmd
	Keys(pattern string) *redis.StringSliceCmd
	Migrate(host, port, key string, db int64, timeout time.Duration) *redis.StatusCmd
	Move(key string, db int64) *redis.BoolCmd
	ObjectRefCount(keys ...string) *redis.IntCmd
	ObjectEncoding(keys ...string) *redis.StringCmd
	ObjectIdleTime(keys ...string) *redis.DurationCmd
	Persist(key string) *redis.BoolCmd
	PExpire(key string, expiration time.Duration) *redis.BoolCmd
	PExpireAt(key string, tm time.Time) *redis.BoolCmd
	PTTL(key string) *redis.DurationCmd
	RandomKey() *redis.StringCmd
	Rename(key, newkey string) *redis.StatusCmd
	RenameNX(key, newkey string) *redis.BoolCmd
	Restore(key string, ttl time.Duration, value string) *redis.StatusCmd
	RestoreReplace(key string, ttl time.Duration, value string) *redis.StatusCmd
	Sort(key string, sort redis.Sort) *redis.StringSliceCmd
	SortInterfaces(key string, sort redis.Sort) *redis.SliceCmd
	TTL(key string) *redis.DurationCmd
	Type(key string) *redis.StatusCmd
	Scan(cursor int64, match string, count int64) *redis.ScanCmd
	SScan(key string, cursor int64, match string, count int64) *redis.ScanCmd
	HScan(key string, cursor int64, match string, count int64) *redis.ScanCmd
	ZScan(key string, cursor int64, match string, count int64) *redis.ScanCmd
	Append(key, value string) *redis.IntCmd
	BitCount(key string, bitCount *redis.BitCount) *redis.IntCmd
	BitOpAnd(destKey string, keys ...string) *redis.IntCmd
	BitOpOr(destKey string, keys ...string) *redis.IntCmd
	BitOpXor(destKey string, keys ...string) *redis.IntCmd
	BitOpNot(destKey string, key string) *redis.IntCmd
	BitPos(key string, bit int64, pos ...int64) *redis.IntCmd
	Decr(key string) *redis.IntCmd
	DecrBy(key string, decrement int64) *redis.IntCmd
	Get(key string) *redis.StringCmd
	GetBit(key string, offset int64) *redis.IntCmd
	GetRange(key string, start, end int64) *redis.StringCmd
	GetSet(key string, value interface{}) *redis.StringCmd
	Incr(key string) *redis.IntCmd
	IncrBy(key string, value int64) *redis.IntCmd
	IncrByFloat(key string, value float64) *redis.FloatCmd
	MGet(keys ...string) *redis.SliceCmd
	MSet(pairs ...string) *redis.StatusCmd
	MSetNX(pairs ...string) *redis.BoolCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetBit(key string, offset int64, value int) *redis.IntCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetXX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	SetRange(key string, offset int64, value string) *redis.IntCmd
	StrLen(key string) *redis.IntCmd
	HDel(key string, fields ...string) *redis.IntCmd
	HExists(key, field string) *redis.BoolCmd
	HGet(key, field string) *redis.StringCmd
	HGetAll(key string) *redis.StringSliceCmd
	HGetAllMap(key string) *redis.StringStringMapCmd
	HIncrBy(key, field string, incr int64) *redis.IntCmd
	HIncrByFloat(key, field string, incr float64) *redis.FloatCmd
	HKeys(key string) *redis.StringSliceCmd
	HLen(key string) *redis.IntCmd
	HMGet(key string, fields ...string) *redis.SliceCmd
	HMSet(key, field, value string, pairs ...string) *redis.StatusCmd
	HMSetMap(key string, fields map[string]string) *redis.StatusCmd
	HSet(key, field, value string) *redis.BoolCmd
	HSetNX(key, field, value string) *redis.BoolCmd
	HVals(key string) *redis.StringSliceCmd
	BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd
	BRPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd
	BRPopLPush(source, destination string, timeout time.Duration) *redis.StringCmd
	LIndex(key string, index int64) *redis.StringCmd
	LInsert(key, op, pivot, value string) *redis.IntCmd
	LLen(key string) *redis.IntCmd
	LPop(key string) *redis.StringCmd
	LPush(key string, values ...string) *redis.IntCmd
	LPushX(key, value interface{}) *redis.IntCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	LRem(key string, count int64, value interface{}) *redis.IntCmd
	LSet(key string, index int64, value interface{}) *redis.StatusCmd
	LTrim(key string, start, stop int64) *redis.StatusCmd
	RPop(key string) *redis.StringCmd
	RPopLPush(source, destination string) *redis.StringCmd
	RPush(key string, values ...string) *redis.IntCmd
	RPushX(key string, value interface{}) *redis.IntCmd
	SAdd(key string, members ...string) *redis.IntCmd
	SCard(key string) *redis.IntCmd
	SDiff(keys ...string) *redis.StringSliceCmd
	SDiffStore(destination string, keys ...string) *redis.IntCmd
	SInter(keys ...string) *redis.StringSliceCmd
	SInterStore(destination string, keys ...string) *redis.IntCmd
	SIsMember(key string, member interface{}) *redis.BoolCmd
	SMembers(key string) *redis.StringSliceCmd
	SMove(source, destination string, member interface{}) *redis.BoolCmd
	SPop(key string) *redis.StringCmd
	SRandMember(key string) *redis.StringCmd
	SRandMemberN(key string, count int64) *redis.StringSliceCmd
	SRem(key string, members ...string) *redis.IntCmd
	SUnion(keys ...string) *redis.StringSliceCmd
	SUnionStore(destination string, keys ...string) *redis.IntCmd
	ZAdd(key string, members ...redis.Z) *redis.IntCmd
	ZAddNX(key string, members ...redis.Z) *redis.IntCmd
	ZAddXX(key string, members ...redis.Z) *redis.IntCmd
	ZAddCh(key string, members ...redis.Z) *redis.IntCmd
	ZAddNXCh(key string, members ...redis.Z) *redis.IntCmd
	ZAddXXCh(key string, members ...redis.Z) *redis.IntCmd
	ZIncr(key string, member redis.Z) *redis.FloatCmd
	ZIncrNX(key string, member redis.Z) *redis.FloatCmd
	ZIncrXX(key string, member redis.Z) *redis.FloatCmd
	ZCard(key string) *redis.IntCmd
	ZCount(key, min, max string) *redis.IntCmd
	ZIncrBy(key string, increment float64, member string) *redis.FloatCmd
	ZInterStore(destination string, store redis.ZStore, keys ...string) *redis.IntCmd
	ZRange(key string, start, stop int64) *redis.StringSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRangeByScore(key string, opt redis.ZRangeByScore) *redis.StringSliceCmd
	ZRangeByLex(key string, opt redis.ZRangeByScore) *redis.StringSliceCmd
	ZRangeByScoreWithScores(key string, opt redis.ZRangeByScore) *redis.ZSliceCmd
	ZRank(key, member string) *redis.IntCmd
	ZRem(key string, members ...string) *redis.IntCmd
	ZRemRangeByRank(key string, start, stop int64) *redis.IntCmd
	ZRemRangeByScore(key, min, max string) *redis.IntCmd
	ZRevRange(key string, start, stop int64) *redis.StringSliceCmd
	ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRevRangeByScore(key string, opt redis.ZRangeByScore) *redis.StringSliceCmd
	ZRevRangeByLex(key string, opt redis.ZRangeByScore) *redis.StringSliceCmd
	ZRevRangeByScoreWithScores(key string, opt redis.ZRangeByScore) *redis.ZSliceCmd
	ZRevRank(key, member string) *redis.IntCmd
	ZScore(key, member string) *redis.FloatCmd
	ZUnionStore(dest string, store redis.ZStore, keys ...string) *redis.IntCmd
	PFAdd(key string, fields ...string) *redis.IntCmd
	PFCount(keys ...string) *redis.IntCmd
	PFMerge(dest string, keys ...string) *redis.StatusCmd
	BgRewriteAOF() *redis.StatusCmd
	BgSave() *redis.StatusCmd
	ClientKill(ipPort string) *redis.StatusCmd
	ClientList() *redis.StringCmd
	ClientPause(dur time.Duration) *redis.BoolCmd
	ClientSetName(name string) *redis.BoolCmd
	//ClientGetName() *redis.StringCmd
	ConfigGet(parameter string) *redis.SliceCmd
	ConfigResetStat() *redis.StatusCmd
	ConfigSet(parameter, value string) *redis.StatusCmd
	DbSize() *redis.IntCmd
	FlushAll() *redis.StatusCmd
	FlushDb() *redis.StatusCmd
	Info(section ...string) *redis.StringCmd
	LastSave() *redis.IntCmd
	Save() *redis.StatusCmd
	Shutdown() *redis.StatusCmd
	ShutdownSave() *redis.StatusCmd
	ShutdownNoSave() *redis.StatusCmd
	SlaveOf(host, port string) *redis.StatusCmd
	Time() *redis.StringSliceCmd
	//SlowLog()
	//Sync()
}

func NewRedisService(config *RedisConfig) (RedisService, base.Error) {
	var client RedisService
	if config.Cluster {
		clusterConfig := &redis.ClusterOptions{
			Addrs:              config.Addrs,
			MaxRedirects:       config.MaxRedirects,
			Password:           config.Password,
			DialTimeout:        config.DialTimeout,
			ReadTimeout:        config.ReadTimeout,
			WriteTimeout:       config.WriteTimeout,
			PoolSize:           config.PoolSize,
			PoolTimeout:        config.PoolTimeout,
			IdleTimeout:        config.IdleTimeout,
			IdleCheckFrequency: config.IdleCheckFrequency,
		}
		client = redis.NewClusterClient(clusterConfig)
	} else {
		opt := &redis.Options{
			Addr:               config.Addrs[0],
			MaxRetries:         config.MaxRedirects,
			Password:           config.Password,
			DialTimeout:        config.DialTimeout,
			ReadTimeout:        config.ReadTimeout,
			WriteTimeout:       config.WriteTimeout,
			PoolSize:           config.PoolSize,
			PoolTimeout:        config.PoolTimeout,
			IdleTimeout:        config.IdleTimeout,
			IdleCheckFrequency: config.IdleCheckFrequency,
		}
		client = redis.NewClient(opt)
	}
	err := client.Ping().Err()
	if err != nil {
		return nil, base.NewError(base.ERROR_CODE_BASE_INIT_ERROR, fmt.Sprintf("Redis Ping 失败:%s", err))
	}
	return client, nil
	//return &_RedisService{
	//	client:client,
	//},nil
}

type _RedisService struct {
	client redis.ClusterClient
}
