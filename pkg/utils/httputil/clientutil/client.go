// Copyright 2023 The kubegems.io Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientutil

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

type BaseClient struct {
	Server              string
	CompleteRequestFunc func(req *http.Request)
	ErrorDecodeFunc     func(resp *http.Response) error
	DataDecoderWrapper  func(data any) any
}

func (c *BaseClient) Request(ctx context.Context, method string, path string, queries map[string]string, data interface{}, into interface{}) error {
	var body io.Reader

	switch typed := data.(type) {
	case []byte:
		body = bytes.NewReader(typed)
	case nil:
	default:
		bts, err := json.Marshal(typed)
		if err != nil {
			return err
		}
		body = bytes.NewReader(bts)
	}
	if len(queries) != 0 {
		vals := url.Values{}
		for k, v := range queries {
			vals.Set(k, v)
		}
		path += "?" + vals.Encode()
	}
	req, err := http.NewRequest(method, c.Server+path, body)
	if err != nil {
		return err
	}
	if c.CompleteRequestFunc != nil {
		c.CompleteRequestFunc(req)
	}
	if req.Header.Get("Content-Type") == "" {
		req.Header.Add("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// not 200~
	// nolint: nestif
	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusIMUsed {
		if c.ErrorDecodeFunc != nil {
			return c.ErrorDecodeFunc(resp)
		}
		// nolint: gomnd
		errmsg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return errors.New(string(errmsg))
	}
	if into == nil {
		return nil
	}

	if c.DataDecoderWrapper != nil {
		into = c.DataDecoderWrapper(into)
	}
	return json.NewDecoder(resp.Body).Decode(into)
}
