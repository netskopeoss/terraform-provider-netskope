package hooks

import (
	"bytes"
	"io"
	"net/http"
)

// rewriteRequestJSON reads req.Body, passes the raw bytes to fn, and installs
// fn's output as the new body with a corrected Content-Length.
// fn receives the raw bytes so callers choose their own decode target.
// Errors from fn are returned unchanged.
func rewriteRequestJSON(req *http.Request, fn func([]byte) ([]byte, error)) (*http.Request, error) {
	if req.Body == nil {
		return req, nil
	}
	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	out, err := fn(body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(out))
	req.ContentLength = int64(len(out))
	return req, nil
}

// rewriteResponseJSON reads res.Body, passes the raw bytes to fn, and installs
// fn's output via replaceBody (which also fixes Content-Type and Content-Length).
// Errors from fn are returned unchanged.
func rewriteResponseJSON(res *http.Response, fn func([]byte) ([]byte, error)) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	out, err := fn(body)
	if err != nil {
		return nil, err
	}
	return replaceBody(res, out), nil
}
