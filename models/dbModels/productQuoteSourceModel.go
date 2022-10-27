package dbModels

type ProductType int

const (
	ProductType_None ProductType = iota
	ProductType_Stock
	ProductType_Crypto
	ProductType_Forex
	ProductType_Futures
)

type ProductQuoteSourceModel struct {
	ProductID    uint64      `gorm:"column:id; primary_key"`
	Type         ProductType `gorm:"column:type"`
	SourceCode   string      `gorm:"column:source_code"`
	QuoteCode    string      `gorm:"column:quote_code"`
	CurrencyCode string      `gorm:"column:currency_code"`
	Interval     string      `gorm:"column:interval"` // HHMMSS
	Status       int         `gorm:"column:status"`   // 1:enabled , 2:disabled
}
