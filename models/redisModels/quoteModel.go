package redisModels

type ProductType int

const (
	ProductType_None ProductType = iota
	ProductType_Stock
	ProductType_Crypto
	ProductType_Forex
	ProductType_Futures
)

type QuoteModel struct {
	ProductID    uint64 `redis:"product_id"`
	Type         ProductType
	SourceCode   string
	QuoteCode    string
	CurrencyCode string
	Quote        map[string]string
	// Ask          decimal.Decimal
	// Bid          decimal.Decimal
	// Open         decimal.Decimal
	// Close        decimal.Decimal
	// High         decimal.Decimal
	// Low          decimal.Decimal
	// Volume       decimal.Decimal
}
