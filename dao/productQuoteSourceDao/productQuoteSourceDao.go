package productQuoteSourceDao

import (
	"errors"

	"github.com/paper-trade-chatbot/be-quote/models/dbModels"

	"gorm.io/gorm"
)

const table = "product_quote_source"

// QueryModel set query condition, used by queryChain()
type QueryModel struct {
	ProductID  uint64
	ProductIDs []uint64
	SourceCode string
}

// New a row
func New(tx *gorm.DB, model *dbModels.ProductQuoteSourceModel) (uint64, error) {

	err := tx.Table(table).
		Create(model).Error

	if err != nil {
		return 0, err
	}
	return model.ProductID, nil
}

// New rows
func News(db *gorm.DB, models []*dbModels.ProductQuoteSourceModel) (int, error) {

	err := db.Transaction(func(tx *gorm.DB) error {

		err := tx.Table(table).
			CreateInBatches(models, 3000).Error

		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(models), nil
}

// Get return a record as raw-data-form
func Get(tx *gorm.DB, query *QueryModel) (*dbModels.ProductQuoteSourceModel, error) {

	result := &dbModels.ProductQuoteSourceModel{}
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

// Gets return records as raw-data-form
func Gets(tx *gorm.DB, query *QueryModel) ([]*dbModels.ProductQuoteSourceModel, error) {
	result := make([]*dbModels.ProductQuoteSourceModel, 0)
	err := tx.Table(table).
		Scopes(queryChain(query)).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return []*dbModels.ProductQuoteSourceModel{}, nil
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func queryChain(query *QueryModel) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.
			Scopes(productIDEqualScope(query.ProductID)).
			Scopes(sourceCodeEqualScope(query.SourceCode)).
			Scopes(productIDInScope(query.ProductIDs))
	}
}

func productIDEqualScope(id uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if id != 0 {
			return db.Where(table+".product_id = ?", id)
		}
		return db
	}
}

func productIDInScope(ids []uint64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(ids) > 0 {
			return db.Where(table+".product_id IN ?", ids)
		}
		return db
	}
}

func sourceCodeEqualScope(code string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if code != "" {
			return db.Where(table+".source_code = ?", code)
		}
		return db
	}
}
