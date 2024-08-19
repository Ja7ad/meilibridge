package types

import "errors"

type TriggerOpType string

const (
	INSERT TriggerOpType = "INSERT"
	UPDATE TriggerOpType = "UPDATE"
	DELETE TriggerOpType = "DELETE"
)

type TriggerRequestBody struct {
	IndexUID string        `json:"index_uid"`
	Type     TriggerOpType `json:"type"`
	Document *Document     `json:"document"`
}

type Document struct {
	PrimaryKey   string `json:"primary_key"`
	PrimaryValue any    `json:"primary_value"`
}

func (t *TriggerRequestBody) Validate() error {
	if len(t.IndexUID) == 0 {
		return errors.New("index_uid is empty")
	}

	switch t.Type {
	case INSERT, UPDATE, DELETE:
	default:
		return errors.New("unknown trigger type")
	}

	if t.Document == nil {
		return errors.New("document is empty")
	}

	if len(t.Document.PrimaryKey) == 0 {
		return errors.New("document primary_key is empty")
	}

	if t.Document.PrimaryValue == nil {
		return errors.New("document primary_value is empty")
	}

	return nil
}
