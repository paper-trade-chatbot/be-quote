package cronjob

import (
	"context"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/go-redis/redis/v9"
	"github.com/paper-trade-chatbot/be-common/cache"
	"github.com/paper-trade-chatbot/be-common/logging"
	"github.com/paper-trade-chatbot/be-quote/cronjob/TWSE"
	"github.com/paper-trade-chatbot/be-quote/cronjob/cleanQuotes"
)

func Cron() {

	scheduler := gocron.NewScheduler(time.UTC)

	startTime := time.Now().Truncate(5 * time.Second)
	scheduler.Every(5).Second().StartAt(startTime).Do(work, TWSE.GetTWSEQuote, TWSE.GetTWSEQuoteKey, time.Second*5)

	startTime = time.Now().Truncate(time.Hour)
	scheduler.Every(1).Hour().StartAt(startTime).Do(work, cleanQuotes.CleanQuotes, cleanQuotes.CleanQuotesKey, time.Second*10)

	// Start all the pending jobs
	scheduler.StartAsync()

}

func work(cronjob func(context.Context) error, generateKey func() string, maxDuration time.Duration) {

	// Generate hash object.
	hash := fnv.New64a()

	// Use time as hash component.
	currentTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(currentTimeBytes,
		uint64(time.Now().UnixNano()))

	// Compute hash value.
	hash.Write([]byte(generateKey()))
	hash.Write(currentTimeBytes)

	cronjobID := fmt.Sprintf("%012x", hash.Sum64())[:12]

	ctx := context.WithValue(context.Background(), logging.ContextKeyRequestId, cronjobID)

	funcName := strings.Split(runtime.FuncForPC(reflect.ValueOf(cronjob).Pointer()).Name(), "/")
	logging.Info(ctx, "[cronjob] start %s", funcName[len(funcName)-1])
	key := "cronjob:" + generateKey()

	r, _ := cache.GetRedis()
	if flag, _ := r.SetNX(ctx, key, cronjobID, maxDuration).Result(); !flag {
		logging.Info(ctx, "[Cronjob] key already exist: %s", key)
		return
	}

	ch := make(chan int, 1)

	ctxTimeout, cancel := context.WithTimeout(ctx, maxDuration)
	defer cancel()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Record the stack trace to logging service, or if we cannot
				// find a logging from this request, use the static logging.
				logging.Error(ctx, "\x1b[31m%v\n[Stack Trace]\n%s\x1b[m", r, debug.Stack())
			}
			ch <- 1
		}()
		err := cronjob(ctxTimeout)
		if err != nil {
			logging.Error(ctxTimeout, "[Cronjob] %s error: %v", key, err)
		}
	}()

	select {
	case <-ctxTimeout.Done():
		logging.Error(ctxTimeout, "[Cronjob] %s timeout error: %v", key, ctxTimeout.Err())
	case <-ch:

	}

	value, err := r.Get(ctx, key).Result()
	if err != nil && err.Error() != redis.Nil.Error() && value == cronjobID {
		if err := r.Del(ctx, key).Err(); err != nil && err.Error() != redis.Nil.Error() {
			logging.Error(ctxTimeout, "[Cronjob] %s failed to delete key: %v", key, err)
		}
	}
}
