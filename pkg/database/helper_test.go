package database

import (
	"testing"
	"time"

	"github.com/Ja7ad/meilibridge/config"
	"github.com/stretchr/testify/assert"
)

func Test_DSNMaker(t *testing.T) {
	tests := []struct {
		Name     string
		Source   *config.Database
		Excepted string
	}{
		{
			Name: "mongodb",
			Source: &config.Database{
				Engine:   config.MONGO,
				Host:     "127.0.0.1",
				Port:     27017,
				User:     "root",
				Password: "foobar",
				Database: "test",
			},
			Excepted: "mongodb://root:foobar@127.0.0.1:27017/",
		},
		{
			Name: "mysql",
			Source: &config.Database{
				Engine:   config.MYSQL,
				Host:     "127.0.0.1",
				Port:     3306,
				User:     "root",
				Password: "foobar",
				Database: "test",
			},
			Excepted: "root:foobar@tcp(127.0.0.1:3306)/test",
		},
		{
			Name: "postgres",
			Source: &config.Database{
				Engine:   config.POSTGRES,
				Host:     "127.0.0.1",
				Port:     9920,
				User:     "root",
				Password: "foobar",
				Database: "test",
			},
			Excepted: "host=127.0.0.1 user=root password=foobar dbname=test port=9920",
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			dsn := dsnMaker(tt.Source)
			assert.Equal(t, dsn, tt.Excepted)
		})
	}
}

func TestMapToResult(t *testing.T) {
	testCases := []struct {
		input    map[string]interface{}
		expected Result
	}{
		{
			input: map[string]interface{}{
				"string":      "test",
				"bytes":       []byte("test"),
				"float64":     float64(1.23),
				"float32":     float32(1.23),
				"int":         123,
				"int8":        int8(123),
				"int16":       int16(123),
				"int32":       int32(123),
				"int64":       int64(123),
				"uint":        uint(123),
				"uint8":       uint8(123),
				"uint16":      uint16(123),
				"uint32":      uint32(123),
				"uint64":      uint64(123),
				"date":        []byte("2023-01-01"),
				"datetime":    []byte("2023-01-01 12:34:56"),
				"timestamp":   []byte("2023-01-01 12:34:56.123456"),
				"invalidTime": []byte("invalid"),
			},
			expected: Result{
				"string":      "test",
				"bytes":       "test",
				"float64":     float64(1.23),
				"float32":     float32(1.23),
				"int":         123,
				"int8":        int8(123),
				"int16":       int16(123),
				"int32":       int32(123),
				"int64":       int64(123),
				"uint":        uint(123),
				"uint8":       uint8(123),
				"uint16":      uint16(123),
				"uint32":      uint32(123),
				"uint64":      uint64(123),
				"date":        parseTime("2023-01-01"),
				"datetime":    parseTime("2023-01-01 12:34:56"),
				"timestamp":   parseTime("2023-01-01 12:34:56.123456"),
				"invalidTime": "invalid",
			},
		},
	}

	for _, tc := range testCases {
		result := mapToResult(tc.input)
		for k, v := range tc.expected {
			assert.Equal(t, result[k], v)
		}
	}
}

func parseTime(value string) time.Time {
	t, _ := parseSQLTime(value)
	return t
}

func TestParseSQLTime(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Time
		err      bool
	}{
		{"2023-01-01 12:34:56", parseTime("2023-01-01 12:34:56"), false},
		{"2023-01-01 12:34:56.123456", parseTime("2023-01-01 12:34:56.123456"), false},
		{"2023-01-01 12:34:56.123456789", parseTime("2023-01-01 12:34:56.123456789"), false},
		{"2023-01-01 12:34:56Z", parseTime("2023-01-01 12:34:56Z"), false},
		{"2023-01-01 12:34:56.123456Z", parseTime("2023-01-01 12:34:56.123456Z"), false},
		{"2023-01-01 12:34:56.123456789Z", parseTime("2023-01-01 12:34:56.123456789Z"), false},
		{"2023-01-01", parseTime("2023-01-01"), false},
		{"15:04:05", parseTime("15:04:05"), false},
		{"15:04:05.123456", parseTime("15:04:05.123456"), false},
		{"15:04:05.123456789", parseTime("15:04:05.123456789"), false},
		{"invalid", time.Time{}, true},
	}

	for _, tc := range testCases {
		result, err := parseSQLTime(tc.input)
		if tc.err && err == nil {
			t.Errorf("expected error for input %s, but got none", tc.input)
		}
		if !tc.err && err != nil {
			t.Errorf("expected no error for input %s, but got %v", tc.input, err)
		}
		if !tc.err && !result.Equal(tc.expected) {
			t.Errorf("expected %v for input %s, but got %v", tc.expected, tc.input, result)
		}
	}
}
