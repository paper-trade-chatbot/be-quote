package quoteSourceDao

import (
	"errors"

	"github.com/paper-trade-chatbot/be-quote/models/dbModels"

	"gorm.io/gorm"
)

const table = "quote_source"

// QueryModel set query condition, used by queryChain()
type QueryModel struct {
	ID   uint64
	Code string
}

// New a row
func New(tx *gorm.DB, model *dbModels.QuoteSourceModel) (uint64, error) {

	err := tx.Table(table).
		Create(model).Error

	if err != nil {
		return 0, err
	}
	return model.ID, nil
}

// Get return a record as raw-data-form
func Get(tx *gorm.DB, query *QueryModel) (*dbModels.QuoteSourceModel, error) {

	result := &dbModels.QuoteSourceModel{}
	err := tx.Table(table).
		Scopes(queryChain(query)).
		Scan(result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func queryChain(query *QueryModel) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Scopes(idEqualScope(query.ID)).
			Scopes(codeEqualScope(query.Code))

	}
}

func idEqualScope(id uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if id != 0 {
			return db.Where(table+".id = ?", id)
		}
		return db
	}
}

func codeEqualScope(code string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code != "" {
			return db.Where(table+".code = ?", code)
		}
		return db
	}
}
