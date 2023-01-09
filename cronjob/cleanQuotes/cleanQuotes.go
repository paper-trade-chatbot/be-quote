package cleanQuotes

import (
	"context"
	"strconv"
	"time"

	"github.com/paper-trade-chatbot/be-common/cache"
	"github.com/paper-trade-chatbot/be-common/database"
	"github.com/paper-trade-chatbot/be-common/logging"
	"github.com/paper-trade-chatbot/be-quote/dao/productQuoteSourceDao"
	"github.com/paper-trade-chatbot/be-quote/dao/quoteDao"
)

func CleanQuotes(ctx context.Context) error {

	now := time.Now()
	db := database.GetDB()

	rds, err := cache.GetRedis()
	if err != nil {
		return err
	}

	models, err := productQuoteSourceDao.Gets(db, &productQuoteSourceDao.QueryModel{})
	if err != nil {
		return err
	}

	queries := []*quoteDao.QueryModel{}

	for _, m := range models {
		q := &quoteDao.QueryModel{
			ProductID: m.ProductID,
		}
		queries = append(queries, q)
	}

	quotes, err := quoteDao.Gets(ctx, rds, queries)
	if err != nil {
		logging.Error(ctx, "[CleanQuotes] get redis error: %v", err)
		return err
	}

	lastHour := now.Add(time.Hour * -1).Format("15")
	thisHour := now.Format("15")
	for _, q := range quotes {
		for k := range q.Quote {

			if k[0:2] != lastHour && k[0:2] != thisHour {
				continue
			}

			delete(q.Quote, k)
		}
	}

	err = quoteDao.Delete(ctx, rds, quotes)
	if err != nil {
		logging.Error(ctx, "[CleanQuotes] delete redis error: %v", err)
		return err
	}

	return nil
}

func CleanQuotesKey() string {
	now := time.Now()
	key := "CleanQuotes:" + strconv.Itoa(now.Hour())
	return key
}
