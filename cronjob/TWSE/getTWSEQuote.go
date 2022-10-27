package TWSE

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-trade-chatbot/be-quote/cache"
	"github.com/paper-trade-chatbot/be-quote/dao/productQuoteSourceDao"
	"github.com/paper-trade-chatbot/be-quote/dao/quoteDao"
	"github.com/paper-trade-chatbot/be-quote/database"
	"github.com/paper-trade-chatbot/be-quote/logging"
	"github.com/paper-trade-chatbot/be-quote/models/dbModels"
	"github.com/paper-trade-chatbot/be-quote/models/redisModels"
)

type TWSEQuote struct {
}

func GetTWSEQuote(ctx context.Context) error {

	now := time.Now()

	db := database.GetDB()

	rds, err := cache.GetRedis()
	if err != nil {
		return err
	}

	models, err := productQuoteSourceDao.Gets(db, &productQuoteSourceDao.QueryModel{
		SourceCode: "TWSE",
	})
	if err != nil {
		return err
	}

	quoteString := "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?json=1&delay=0&ex_ch="

	for _, m := range models {
		quoteString += m.QuoteCode + "|"
	}

	resp, err := http.Get(quoteString)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}

	modelsMap := map[string]*dbModels.ProductQuoteSourceModel{}
	for _, m := range models {
		modelsMap[m.QuoteCode] = m
	}

	if _, ok := result["msgArray"]; !ok {
		return errors.New("no message array")
	}
	if _, ok := result["msgArray"].(map[string]interface{}); !ok {
		return errors.New("message array cast failed")
	}

	quotes := []*redisModels.QuoteModel{}

	for _, msg := range result["msgArray"].([]map[string]interface{}) {

		if _, ok := modelsMap[msg["ch"].(string)]; !ok {
			logging.Warn(ctx, "[GetTWSEQuote] no such code in models: %s", msg["ch"].(string))
			continue
		}

		quoteTime := now.Format("150405")

		quote := &redisModels.QuoteModel{
			ProductID:    modelsMap[msg["ch"].(string)].ProductID,
			Type:         redisModels.ProductType(modelsMap[msg["ch"].(string)].Type),
			SourceCode:   modelsMap[msg["ch"].(string)].SourceCode,
			QuoteCode:    modelsMap[msg["ch"].(string)].QuoteCode,
			CurrencyCode: modelsMap[msg["ch"].(string)].CurrencyCode,
			Quote: map[string]string{
				"ask":     msg["a"].(string),
				"bid":     msg["b"].(string),
				quoteTime: msg["z"].(string),
			},
		}

		quotes = append(quotes, quote)
	}

	err = quoteDao.Updates(ctx, rds, quotes)
	if err != nil {
		logging.Error(ctx, "[GetTWSEQuote] update redis error: %v", err)
		return err
	}

	return nil
}

func GetTWSEQuoteKey() string {
	now := time.Now()
	key := "getTWSEQuote:" + strconv.Itoa(now.Hour()) + "-" + strconv.Itoa(now.Minute()) + "-" + strconv.Itoa(now.Second())
	return key
}