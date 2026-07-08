package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// publisherDeleteGuardHook checks for connected apps before allowing a publisher
// DELETE to proceed. The Netskope API rejects deletes for publishers with connected
// apps but returns a generic error. This hook issues a GET before the DELETE and
// returns a descriptive error listing the connected app names if any are present.
//
// If the pre-check GET fails for any reason the hook passes through and lets the
// DELETE proceed — better to surface a raw API error than to silently block a
// valid delete.
//
// Connected-apps shape note (BUG-017): the GET-by-ID endpoint returns
// connected_apps as an array of objects; the list endpoint returns strings.
// This hook handles both shapes via extractConnectedAppNames.
type publisherDeleteGuardHook struct{}

var _ beforeRequestHook = (*publisherDeleteGuardHook)(nil)

func (h *publisherDeleteGuardHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "deleteNPAPublishers" {
		return req, nil
	}

	client := hookCtx.SDKConfiguration.Client
	if client == nil {
		return req, nil
	}

	ctx := hookCtx.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// The DELETE URL already contains the publisher ID — issue a GET to the same path.
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL.String(), nil)
	if err != nil {
		return req, nil
	}

	// Copy auth headers from the outgoing DELETE request.
	for _, hdr := range []string{"Netskope-Api-Token", "Authorization"} {
		if val := req.Header.Get(hdr); val != "" {
			getReq.Header.Set(hdr, val)
		}
	}

	resp, err := client.Do(getReq)
	if err != nil || resp.StatusCode != http.StatusOK {
		return req, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return req, nil
	}

	var payload struct {
		Data struct {
			Name          string            `json:"name"`
			ConnectedApps []json.RawMessage `json:"connected_apps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return req, nil
	}

	names := extractConnectedAppNames(payload.Data.ConnectedApps)
	if len(names) == 0 {
		return req, nil
	}

	publisherName := payload.Data.Name
	if publisherName == "" {
		parts := strings.Split(strings.TrimRight(req.URL.Path, "/"), "/")
		publisherName = parts[len(parts)-1]
	}

	return nil, fmt.Errorf(
		"cannot delete publisher %q: %d app(s) still connected:\n  - %s\nDetach or delete these apps before destroying the publisher",
		publisherName,
		len(names),
		strings.Join(names, "\n  - "),
	)
}

// extractConnectedAppNames pulls app name strings from connected_apps elements,
// which may be JSON strings or JSON objects depending on the API endpoint (BUG-017).
func extractConnectedAppNames(elems []json.RawMessage) []string {
	names := make([]string, 0, len(elems))
	for _, elem := range elems {
		if len(elem) == 0 {
			continue
		}
		switch elem[0] {
		case '"':
			// String form from list endpoint: "[AppName]"
			var s string
			if json.Unmarshal(elem, &s) == nil && s != "" {
				names = append(names, s)
			}
		case '{':
			// Object form from get-by-ID endpoint: {"name": "[AppName]", ...}
			var obj struct {
				Name string `json:"name"`
			}
			if json.Unmarshal(elem, &obj) == nil && obj.Name != "" {
				names = append(names, obj.Name)
			}
		}
	}
	return names
}
