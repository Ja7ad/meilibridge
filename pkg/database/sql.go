package database

import (
	"context"
	"math"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/Ja7ad/meilibridge/pkg/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type SQL struct {
	db  *gorm.DB
	log logger.Logger
}

func newSQL(
	src *config.Source,
	log logger.Logger,
) (SQLExecutor, error) {
	s := &SQL{
		log: log,
	}

	dsn := dsnMaker(src)

	switch src.Engine {
	case config.MYSQL:
		db, err := gorm.Open(mysql.Open(dsn))
		if err != nil {
			return nil, err
		}
		s.db = db
	case config.POSTGRES:
		db, err := gorm.Open(postgres.Open(dsn))
		if err != nil {
			return nil, err
		}
		s.db = db
	}

	return s, nil
}

func (s *SQL) Close() error {
	sq, err := s.db.DB()
	if err != nil {
		return err
	}
	return sq.Close()
}

func (s *SQL) Count(ctx context.Context, table string) (int64, error) {
	var count int64
	db := s.db.WithContext(ctx).Table(table)
	return count, db.Count(&count).Error
}

func (s *SQL) FindOne(ctx context.Context, table string, query map[string]interface{}) (Result, error) {
	queryStr, args := buildQueryFindOne(table, query)

	rows, err := s.db.WithContext(ctx).Raw(queryStr, args...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		data, err := decodeRows(rows)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	return nil, err
}

func (s *SQL) FindLimit(ctx context.Context, table string, limit int64) (Cursor, error) {
	count, err := s.Count(ctx, table)
	if err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(count) / float64(limit)))

	return &sqlCursor{
		total: count,
		pages: int(totalPages),
		page:  0,
		limit: int(limit),
		db:    s.db,
		table: table,
		err:   nil,
	}, nil
}

type sqlCursor struct {
	total int64
	pages int
	page  int
	limit int
	db    *gorm.DB
	table string
	err   error
	res   []*Result
}

func (c *sqlCursor) Next(ctx context.Context) bool {
	if c.page >= c.pages {
		return false
	}

	if len(c.res) != 0 {
		c.res = make([]*Result, 0)
	}

	skip := c.page * c.limit

	rows, err := c.db.WithContext(ctx).Table(c.table).Offset(skip).Limit(c.limit).Rows()
	if err != nil {
		c.err = err
		return false
	}
	defer rows.Close()

	for i := 0; i < c.limit; i++ {
		for rows.Next() {
			data, err := decodeRows(rows)
			if err != nil {
				c.err = err
				return false
			}
			res := mapToResult(data)
			c.res = append(c.res, &res)

		}
	}

	c.page++
	return true
}

func (c *sqlCursor) Result() ([]*Result, error) {
	return c.res, c.err
}
