package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// 拉取 go-redis v9 依赖
// go get github.com/redis/go-redis/v9
// 整理依赖（推荐）
// go mod tidy

// bitmap 示例代码
func main() {
	// 初始化一个 Redis 客户端，并指定 Redis 服务所在的地址和端口。在后续的操作中，可以通过 rdb 对象与 Redis 服务器进行通信。
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	ctx := context.Background()

	// bitmapDemo(ctx, rdb)
	bitFieldDemo(ctx, rdb)
}

// bitmap 示例
func bitmapDemo(ctx context.Context, rdb *redis.Client) {
	key := "test:bitmap"

	// setbit
	// 把 test:bitmap 这张表里的 第 0 个位置 设成 1
	// SETBIT 返回的是“设置之前这个位置的值”
	// 第一次设置返回值一定是0（之前是没开）
	val1 := rdb.SetBit(ctx, key, 0, 1).Val()
	fmt.Printf("setbit ret: %v\n", val1)

	// getbit
	val2 := rdb.GetBit(ctx, key, 0).Val()
	fmt.Printf("getbit ret: %v\n", val2)

	// bitcount：统计有多少个 bit 是 1
	// Start:0, End:-1：表示从第 0 个字节开始到最后一个字节（-1 是末尾）
	val3 := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: -1}).Val()
	fmt.Printf("bitcount ret: %v\n", val3)

	// bitop
	key2 := "test:bitmap:2"
	rdb.SetBit(ctx, key2, 2, 1)
	// offset: 0 1 2
	// 		   1 0 0
	// 		   0 0 1
	key3 := "test:bitmap:3"

	// BITOP AND：对两个 Bitmap 做按位与，结果写入 key3
	val4 := rdb.BitOpAnd(ctx, key3, key, key2).Val()
	fmt.Printf("bitop ret: %v\n", val4)

	val5 := rdb.BitCount(ctx, key3, &redis.BitCount{Start: 0, End: -1}).Val()
	fmt.Printf("bitcount ret: %v\n", val5)

	// bitpos: 找“第一个等于 1 的 bit 在哪里”
	val7 := rdb.BitPos(ctx, key2, 1).Val()
	fmt.Printf("bitpos ret: %v\n", val7)

	rdb.SetBit(ctx, key2, 0, 1)
	rdb.SetBit(ctx, key2, 1, 1)
	val6 := rdb.BitCount(ctx, key2, &redis.BitCount{Start: 0, End: -1}).Val()
	fmt.Printf("get ret: %v\n", val6)

	// Bitmap 在 Redis 里，本质还是 String，但它不是“给人看的字符串”
	// val8 := rdb.Get(ctx, key2).Val()
	// fmt.Printf("val8: %v\n", val8)

	// 只有在你存的是 普通字符串 / 数字字符串 时：GET 才“合理”
	rdb.Set(ctx, "score", "100", 0)
	val9 := rdb.Get(ctx, "score").Val() // "100"
	fmt.Printf("val9: %v\n", val9)
}

func bitFieldDemo(ctx context.Context, rdb *redis.Client) {
	key := "test:checkin:2026" // 存储2026年整年的打卡情况
	// 假设现在是 2026-05
	// 5.1 打卡
	// 1.1 0
	// 1.2 1
	// 1.3 2
	// ...
	// 12.31 365
	// 计算出5.1号是今年的第几天，索引位就是天数-1
	t, _ := time.Parse("2006-01-02", "2026-05-01")
	offset := t.YearDay() - 1                // 5.1
	offset52 := offset + 1                   // 5.2
	offset53 := offset + 2                   // 5.3
	rdb.SetBit(ctx, key, int64(offset52), 1) // 5.2打卡
	rdb.SetBit(ctx, key, int64(offset53), 1) // 5.3打卡
	// 5.1（bit[120]）：0（没设置它）
	// 5.2（bit[121]）：1
	// 5.3（bit[122]）：1

	// 查看5月的打卡情况
	// 从5.1 到 5.31的数据读取出来
	// 5.1 到 5.31 共多少天?

	// 用多大的整数能存下31天
	// 从 offset=120 开始，读 31 个 bit
	ret := rdb.BitField(ctx, key, "GET", "u31", offset, "GET", "u4", offset).Val()
	fmt.Printf("ret:%v\n", ret)
	fmt.Printf("bet:%031b\n", ret[0])
	fmt.Printf("bet:%04b\n", ret[1])
}
