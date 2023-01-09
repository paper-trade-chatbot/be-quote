package TWSE

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paper-trade-chatbot/be-common/cache"
	"github.com/paper-trade-chatbot/be-common/database"
	"github.com/paper-trade-chatbot/be-common/logging"
	"github.com/paper-trade-chatbot/be-quote/dao/productQuoteSourceDao"
	"github.com/paper-trade-chatbot/be-quote/dao/quoteDao"
	"github.com/paper-trade-chatbot/be-quote/models/dbModels"
	"github.com/paper-trade-chatbot/be-quote/models/redisModels"
)

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
	if _, ok := result["msgArray"].([]interface{}); !ok {
		logging.Error(ctx, "[GetTWSEQuote] message array cast failed. %#v", result["msgArray"])
		return errors.New("message array cast failed")
	}

	quotes := []*redisModels.QuoteModel{}

	for _, m := range result["msgArray"].([]interface{}) {
		msg := m.(map[string]interface{})
		if _, ok := modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)]; !ok {
			logging.Warn(ctx, "[GetTWSEQuote] no such code in models: %s", msg["ch"].(string))
			continue
		}

		quoteTime := now.Format("150405")

		ask := strings.Split(msg["a"].(string), "_")
		bid := strings.Split(msg["b"].(string), "_")

		quote := &redisModels.QuoteModel{
			ProductID:    modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)].ProductID,
			Type:         redisModels.ProductType(modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)].Type),
			SourceCode:   modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)].SourceCode,
			QuoteCode:    modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)].QuoteCode,
			CurrencyCode: modelsMap[msg["ex"].(string)+"_"+msg["ch"].(string)].CurrencyCode,
			Quote: map[string]string{
				"ask": ask[0],
				"bid": bid[0],
			},
		}

		if msg["z"].(string) != "-" {
			quote.Quote[quoteTime] = msg["z"].(string)
			quote.Quote["latest"] = msg["z"].(string)
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
	key := "GetTWSEQuote:" + strconv.Itoa(now.Hour()) + "-" + strconv.Itoa(now.Minute()) + "-" + strconv.Itoa(now.Second())
	return key
}
