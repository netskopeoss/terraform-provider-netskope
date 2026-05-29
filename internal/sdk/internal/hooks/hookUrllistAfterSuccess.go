package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// urllistAfterSuccess handles response transformations for URL list endpoints:
//   - createUrllist: extracts single item from array response, deploys, re-fetches
//   - updateUrllist: deploys, re-fetches to return applied state
//   - deleteUrllist: deploys to apply the deletion
//   - listUrllists: wraps raw array into {"items": [...]}
type urllistAfterSuccess struct{}

var _ afterSuccessHook = (*urllistAfterSuccess)(nil)

func (u *urllistAfterSuccess) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createUrllist":
		return u.handleCreate(res)
	case "updateUrllist":
		return u.handleUpdate(res)
	case "deleteUrllist":
		return u.handleDelete(res)
	case "listUrllists":
		return u.handleList(res)
	default:
		return res, nil
	}
}

// handleCreate extracts the created item from the array response, deploys,
// and re-fetches the item by ID to return the applied state.
func (u *urllistAfterSuccess) handleCreate(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("urllist create hook: failed to read response body: %w", err)
	}

	// API returns [{...}] — extract single item
	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("urllist create hook: failed to unmarshal array response: %w", err)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("urllist create hook: empty array response")
	}

	// Parse the item to get the ID
	var item struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(items[0], &item); err != nil {
		return nil, fmt.Errorf("urllist create hook: failed to parse item ID: %w", err)
	}

	// Deploy pending changes
	if err := u.deploy(res); err != nil {
		return nil, fmt.Errorf("urllist create hook: deploy failed: %w", err)
	}

	// Re-fetch the item to get the applied state
	return u.fetchByID(res, item.ID)
}

// handleUpdate deploys pending changes and re-fetches the item.
func (u *urllistAfterSuccess) handleUpdate(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("urllist update hook: failed to read response body: %w", err)
	}

	// Parse the response to get the ID
	var item struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("urllist update hook: failed to parse item ID: %w", err)
	}

	// Deploy pending changes
	if err := u.deploy(res); err != nil {
		return nil, fmt.Errorf("urllist update hook: deploy failed: %w", err)
	}

	// Re-fetch to get applied state
	return u.fetchByID(res, item.ID)
}

// handleDelete deploys the pending deletion.
func (u *urllistAfterSuccess) handleDelete(res *http.Response) (*http.Response, error) {
	// Drain the response body
	io.ReadAll(res.Body)
	res.Body.Close()

	// Deploy pending changes
	if err := u.deploy(res); err != nil {
		return nil, fmt.Errorf("urllist delete hook: deploy failed: %w", err)
	}

	// Return an empty body — Terraform doesn't need the response for delete
	res.Body = io.NopCloser(bytes.NewReader([]byte("{}")))
	res.ContentLength = 2
	return res, nil
}

// handleList wraps the raw array response into {"items": [...]}.
func (u *urllistAfterSuccess) handleList(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("urllist list hook: failed to read response body: %w", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("urllist list hook: failed to unmarshal array: %w", err)
	}

	wrapped, err := json.Marshal(map[string]interface{}{
		"items": items,
	})
	if err != nil {
		return nil, fmt.Errorf("urllist list hook: failed to marshal wrapped response: %w", err)
	}

	return replaceBody(res, wrapped), nil
}

// deploy calls POST /policy/urllist/deploy to apply pending changes.
func (u *urllistAfterSuccess) deploy(origRes *http.Response) error {
	deployURL := *origRes.Request.URL
	deployURL.Path = "/api/v2/policy/urllist/deploy"
	deployURL.RawQuery = ""

	deployReq, err := http.NewRequestWithContext(origRes.Request.Context(), "POST", deployURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to build deploy request: %w", err)
	}

	// Copy auth headers
	for _, h := range []string{"Netskope-Api-Token", "Authorization"} {
		if v := origRes.Request.Header.Get(h); v != "" {
			deployReq.Header.Set(h, v)
		}
	}

	client := &http.Client{}
	deployRes, err := client.Do(deployReq)
	if err != nil {
		return fmt.Errorf("deploy request failed: %w", err)
	}
	io.ReadAll(deployRes.Body)
	deployRes.Body.Close()

	if deployRes.StatusCode != 200 {
		return fmt.Errorf("deploy returned status %d", deployRes.StatusCode)
	}

	return nil
}

// fetchByID does a GET /policy/urllist/{id} and replaces the response body.
func (u *urllistAfterSuccess) fetchByID(origRes *http.Response, id int64) (*http.Response, error) {
	getURL := *origRes.Request.URL
	getURL.Path = fmt.Sprintf("/api/v2/policy/urllist/%d", id)
	getURL.RawQuery = ""

	getReq, err := http.NewRequestWithContext(origRes.Request.Context(), "GET", getURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("urllist hook: failed to build GET request: %w", err)
	}

	// Copy auth headers
	for _, h := range []string{"Netskope-Api-Token", "Authorization"} {
		if v := origRes.Request.Header.Get(h); v != "" {
			getReq.Header.Set(h, v)
		}
	}

	client := &http.Client{}
	getRes, err := client.Do(getReq)
	if err != nil {
		return nil, fmt.Errorf("urllist hook: follow-up GET failed: %w", err)
	}

	body, err := io.ReadAll(getRes.Body)
	getRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("urllist hook: failed to read GET response: %w", err)
	}

	// Strip json_version from data to avoid confusing the SDK
	var item map[string]interface{}
	if err := json.Unmarshal(body, &item); err == nil {
		if data, ok := item["data"].(map[string]interface{}); ok {
			delete(data, "json_version")
			if cleaned, err := json.Marshal(item); err == nil {
				body = cleaned
			}
		}
	}

	return replaceBody(origRes, body), nil
}

// idFromPath extracts the numeric ID from the last path segment.
func idFromPath(path string) (int64, error) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty path")
	}
	return strconv.ParseInt(parts[len(parts)-1], 10, 64)
}
