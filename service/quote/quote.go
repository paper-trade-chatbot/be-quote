package quote

import (
	"context"
	"strconv"
	"time"

	common "github.com/paper-trade-chatbot/be-common"
	"github.com/paper-trade-chatbot/be-proto/quote"
	"github.com/paper-trade-chatbot/be-quote/cache"
	"github.com/paper-trade-chatbot/be-quote/dao/productQuoteSourceDao"
	"github.com/paper-trade-chatbot/be-quote/dao/quoteDao"
	"github.com/paper-trade-chatbot/be-quote/database"
	"github.com/paper-trade-chatbot/be-quote/logging"
	"github.com/paper-trade-chatbot/be-quote/models/dbModels"
)

type QuoteIntf interface {
	AddProductQuoteSources(ctx context.Context, in *quote.AddProductQuoteSourcesReq) (*quote.AddProductQuoteSourcesRes, error)
	ModifyProductQuoteSource(ctx context.Context, in *quote.ModifyProductQuoteSourceReq) (*quote.ModifyProductQuoteSourceRes, error)
	GetQuotes(ctx context.Context, in *quote.GetQuotesReq) (*quote.GetQuotesRes, error)
	DeleteQuotes(ctx context.Context, in *quote.DeleteQuotesReq) (*quote.DeleteQuotesRes, error)
}

type QuoteImpl struct {
	QuoteClient quote.QuoteServiceClient
}

func New() QuoteIntf {
	return &QuoteImpl{}
}

func (impl *QuoteImpl) AddProductQuoteSources(ctx context.Context, in *quote.AddProductQuoteSourcesReq) (*quote.AddProductQuoteSourcesRes, error) {

	db := database.GetDB()

	models := []*dbModels.ProductQuoteSourceModel{}

	for _, p := range in.GetProducts() {
		model := &dbModels.ProductQuoteSourceModel{
			ProductID:    uint64(p.GetProductID()),
			Type:         dbModels.ProductType(p.GetType()),
			SourceCode:   p.GetSourceCode(),
			QuoteCode:    p.GetQuoteCode(),
			CurrencyCode: p.GetCurrencyCode(),
			Status:       int(p.GetStatus()),
		}
		models = append(models, model)
	}

	if len(in.Products) == 0 {
		return nil, common.ErrNoRequiredParam
	}

	if _, err := productQuoteSourceDao.News(db, models); err != nil {
		logging.Error(ctx, "[AddProductQuoteSources] dao news error: %v", err)
		return nil, err
	}

	return &quote.AddProductQuoteSourcesRes{}, nil
}

func (impl *QuoteImpl) ModifyProductQuoteSource(ctx context.Context, in *quote.ModifyProductQuoteSourceReq) (*quote.ModifyProductQuoteSourceRes, error) {

	return nil, common.ErrNotImplemented
}

func (impl *QuoteImpl) GetQuotes(ctx context.Context, in *quote.GetQuotesReq) (*quote.GetQuotesRes, error) {

	rds, err := cache.GetRedis()
	if err != nil {
		return nil, err
	}

	queries := []*quoteDao.QueryModel{}
	for _, p := range in.GetProductIDs() {
		query := &quoteDao.QueryModel{
			ProductID: uint64(p),
		}
		queries = append(queries, query)
	}

	var from, to time.Time
	if in.GetFrom != nil {
		from, err = time.Parse("150405", *in.GetFrom)
		if err != nil {
			logging.Error(ctx, "[GetQuotes] parse time %s error: %v", *in.GetFrom, err)
			return nil, common.ErrInvalidParam
		}
	}
	if in.GetTo != nil {
		to, err = time.Parse("150405", *in.GetTo)
		if err != nil {
			logging.Error(ctx, "[GetQuotes] parse time %s error: %v", *in.GetTo, err)
			return nil, common.ErrInvalidParam
		}
		if *in.GetTo == "000000" {
			to = to.Add(time.Hour * 24)
		}
	}

	models, err := quoteDao.Gets(ctx, rds, queries)
	if err != nil {
		logging.Error(ctx, "[GetQuotes] redis gets error: %v", err)
		return nil, err
	}

	quotes := []*quote.GetQuotesResItem{}

	for _, m := range models {

		resQuote := map[string]string{}

		for k, v := range m.Quote {
			if (k == "ask" && in.Flag&quote.GetQuotesReq_GetFlag_Ask > 0) ||
				(k == "bid" && in.Flag&quote.GetQuotesReq_GetFlag_Bid > 0) ||
				(k == "latest" && in.Flag&quote.GetQuotesReq_GetFlag_Latest > 0) ||
				(in.GetFrom == nil || in.GetTo == nil) {
				resQuote[k] = v
				continue
			}

			if k == "ask" || k == "bid" || k == "latest" {
				continue
			}

			quoteTime, err := time.Parse("150405", k)
			if err != nil {
				logging.Warn(ctx, "[GetQuotes] parse time %s failed: %v", k, err)
				continue
			}
			if k == "000000" {
				quoteTime = quoteTime.Add(time.Hour * 24)
			}

			if quoteTime.After(from) && !quoteTime.After(to) {
				resQuote[k] = v
			}
		}

		q := &quote.GetQuotesResItem{
			ProductID: int64(m.ProductID),
			Quotes:    resQuote,
		}
		quotes = append(quotes, q)
	}

	return &quote.GetQuotesRes{
		Quotes: quotes,
	}, nil
}

func (impl *QuoteImpl) DeleteQuotes(ctx context.Context, in *quote.DeleteQuotesReq) (*quote.DeleteQuotesRes, error) {

	rds, err := cache.GetRedis()
	if err != nil {
		return nil, err
	}

	queries := []*quoteDao.QueryModel{}
	for _, p := range in.GetProductIDs() {
		query := &quoteDao.QueryModel{
			ProductID: uint64(p),
		}
		queries = append(queries, query)
	}

	models, err := quoteDao.Gets(ctx, rds, queries)
	if err != nil {
		logging.Error(ctx, "[GetQuotes] redis gets error: %v", err)
		return nil, err
	}

	var fromHour, fromMinute, fromSecond, toHour, toMinute, toSecond int
	if fromHour, err = strconv.Atoi(in.DeleteFrom[0:2]); err != nil {
		return nil, err
	}
	if fromMinute, err = strconv.Atoi(in.DeleteFrom[2:4]); err != nil {
		return nil, err
	}
	if fromSecond, err = strconv.Atoi(in.DeleteFrom[4:6]); err != nil {
		return nil, err
	}
	if toHour, err = strconv.Atoi(in.DeleteTo[0:2]); err != nil {
		return nil, err
	}
	if toMinute, err = strconv.Atoi(in.DeleteTo[2:4]); err != nil {
		return nil, err
	}
	if toSecond, err = strconv.Atoi(in.DeleteTo[4:6]); err != nil {
		return nil, err
	}
	deleteFrom := time.Date(0, 0, 0, fromHour, fromMinute, fromSecond, 0, time.UTC)
	deleteTo := time.Date(0, 0, 0, toHour, toMinute, toSecond, 0, time.UTC)
	if in.DeleteTo == "000000" {
		deleteTo = time.Date(0, 0, 1, 0, 0, 0, 0, time.UTC)
	}

	for _, m := range models {
		for k := range m.Quote {
			hr, _ := strconv.Atoi(k[0:2])
			min, _ := strconv.Atoi(k[2:4])
			sec, _ := strconv.Atoi(k[4:6])
			keyTime := time.Date(0, 0, 0, hr, min, sec, 0, time.UTC)
			if !keyTime.After(deleteFrom) || keyTime.After(deleteTo) {
				delete(m.Quote, k)
			}
		}
	}

	err = quoteDao.Delete(ctx, rds, models)
	if err != nil {
		logging.Error(ctx, "[GetQuotes] redis delete error: %v", err)
		return nil, err
	}

	return &quote.DeleteQuotesRes{}, nil
}
