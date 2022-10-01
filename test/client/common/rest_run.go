package common

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"

	"github.com/anzx/pkg/log"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Run(ctx context.Context, method string, url string, in proto.Message, headers http.Header) ([]byte, error) {
	body, err := protojson.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header = headers
	log.Info(ctx, "", log.Str("Request", req.URL.String()), log.Any("Headers", req.Header))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 {
		return nil, errors.Errorf("invalid status %d: %s", resp.StatusCode, resp.Status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "", log.Str("Response", req.URL.String()), log.Any("Headers", resp.Header), log.Int("Status", resp.StatusCode), log.Bytes("body", respBody))

	return respBody, nil
}
