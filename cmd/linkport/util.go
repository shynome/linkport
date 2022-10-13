package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func CheckResp(resp *http.Response) (err error) {
	defer err2.Return(&err)
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		errText := try.To1(io.ReadAll(resp.Body))
		err = fmt.Errorf("server err. code: %v. content: %s", resp.StatusCode, errText)
		return
	}
	return
}

func WithTopic(endpoint string, topic string) (rEndpoint string, err error) {
	defer err2.Return(&err)

	if topic == "" {
		return endpoint, nil
	}

	u := try.To1(url.Parse(endpoint))
	q := u.Query()
	q.Set("t", topic)
	u.RawQuery = q.Encode()

	rEndpoint = u.String()

	return
}
