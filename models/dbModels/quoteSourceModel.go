package dbModels

type QuoteSourceModel struct {
	ID      uint64 `gorm:"column:id; primary_key"`
	Code    string `gorm:"column:code"`
	API     string `gorm:"column:api"`
	Example string `gorm:"column:example"`
}
