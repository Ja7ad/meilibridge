package trigger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Ja7ad/meilibridge/pkg/types"
	"net/http"
	"net/url"
)

type Trigger struct {
	client *http.Client
	host   string
	token  string
}

func New(client *http.Client, host, token string) *Trigger {
	return &Trigger{client: client, host: host, token: token}
}

func (t *Trigger) Trigger(ctx context.Context, bridge string, req *types.TriggerRequestBody) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hook, err := url.JoinPath(t.host, bridge, req.IndexUID)
	if err != nil {
		return err
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, hook, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	r.Header.Set("Content-Type", "application/json")
	if len(t.token) != 0 {
		r.Header.Set("x-token-key", t.token)
	}

	resp, err := t.client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		var msg any
		if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
			return err
		}
		return errors.New(fmt.Sprintf("%v", msg))
	}

	return nil
}
