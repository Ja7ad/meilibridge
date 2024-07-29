package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Ja7ad/meilibridge/config"
)

func dsnMaker(src *config.Database) string {
	dsn := strings.Builder{}

	paramsGenerator := func() string {
		p := make([]string, 0)
		for k, v := range src.CustomParams {
			p = append(p, fmt.Sprintf("%s=%v", k, v))
		}

		if src.Engine == config.POSTGRES {
			return strings.Join(p, " ")
		}

		return strings.Join(p, "&")
	}

	switch src.Engine {
	case config.MONGO:
		dsn.WriteString("mongodb://")
		if src.User != "" && src.Password != "" {
			dsn.WriteString(src.User)
			dsn.WriteString(":")
			dsn.WriteString(src.Password)
			dsn.WriteString("@")
		}
		dsn.WriteString(src.Host)
		dsn.WriteString(":")
		dsn.WriteString(strconv.Itoa(int(src.Port)))
		dsn.WriteString("/")
		if len(src.CustomParams) > 0 {
			dsn.WriteString("?")
			dsn.WriteString(paramsGenerator())
		}
	case config.MYSQL:
		if src.User != "" && src.Password != "" {
			dsn.WriteString(src.User)
			dsn.WriteString(":")
			dsn.WriteString(src.Password)
			dsn.WriteString("@")
		}
		dsn.WriteString("tcp(")
		dsn.WriteString(src.Host)
		dsn.WriteString(":")
		dsn.WriteString(strconv.Itoa(int(src.Port)))
		dsn.WriteString(")/")
		dsn.WriteString(src.Database)
		if len(src.CustomParams) > 0 {
			dsn.WriteString("?")
			dsn.WriteString(paramsGenerator())
		}
	case config.POSTGRES:
		dsn.WriteString(fmt.Sprintf("host=%s ", src.Host))
		if src.User != "" && src.Password != "" {
			dsn.WriteString(fmt.Sprintf("user=%s ", src.User))
			dsn.WriteString(fmt.Sprintf("password=%s ", src.Password))
		}
		dsn.WriteString(fmt.Sprintf("dbname=%s ", src.Database))
		dsn.WriteString(fmt.Sprintf("port=%d", int(src.Port)))
		if len(src.CustomParams) > 0 {
			dsn.WriteString(" ")
			dsn.WriteString(paramsGenerator())
		}
	}

	return dsn.String()
}

func buildQueryFindOne(table string, query map[string]interface{}) (string, []interface{}) {
	whereClause, args := buildWhereClause(query)
	queryStr := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", table, whereClause)
	return queryStr, args
}

func buildWhereClause(query map[string]interface{}) (string, []interface{}) {
	whereClause := ""
	args := make([]interface{}, 0, len(query))

	for k, v := range query {
		if whereClause != "" {
			whereClause += " AND "
		}
		whereClause += fmt.Sprintf("%s = ?", k)
		args = append(args, v)
	}

	return whereClause, args
}

func mapToResult(q map[string]interface{}) Result {
	result := make(Result)

	for k, v := range q {
		switch v.(type) {
		case string:
			result[k] = v.(string)
		case []byte:
			strValue := string(v.([]byte))
			if parsedTime, err := parseSQLTime(strValue); err == nil {
				result[k] = parsedTime
			} else {
				result[k] = strValue
			}
		case float64:
			result[k] = v.(float64)
		case float32:
			result[k] = v.(float32)
		case int:
			result[k] = v.(int)
		case int8:
			result[k] = v.(int8)
		case int16:
			result[k] = v.(int16)
		case int32:
			result[k] = v.(int32)
		case int64:
			result[k] = v.(int64)
		case uint:
			result[k] = v.(uint)
		case uint8:
			result[k] = v.(uint8)
		case uint16:
			result[k] = v.(uint16)
		case uint32:
			result[k] = v.(uint32)
		case uint64:
			result[k] = v.(uint64)
		default:
			result[k] = v
		}
	}

	return result
}

func parseSQLTime(value string) (time.Time, error) {
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05Z07:00",
		"2006-01-02 15:04:05.999999Z07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02",
		"15:04:05",
		"15:04:05.999999",
		"15:04:05.999999999",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, value); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse time: %s", value)
}

func decodeRows(rows *sql.Rows) (map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)
	if err != nil {
		return nil, err
	}

	for i, col := range columns {
		data[col] = values[i]
	}

	return data, nil
}
