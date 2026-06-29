package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// aigTokenRequestHook removes expire_in from token create/update requests when
// it was not explicitly configured. The Terraform framework can populate the
// expire_in nested object with zero values (value=0, unit="") when the attribute
// is null in the plan — the API rejects these as invalid.
type aigTokenRequestHook struct{}

var _ beforeRequestHook = (*aigTokenRequestHook)(nil)

func (h *aigTokenRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "createAigToken" && hookCtx.OperationID != "updateAigToken" {
		return req, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("aigTokenRequestHook: unable to read request body: %w", err)
	}

	var requestMap map[string]interface{}
	if err := json.Unmarshal(body, &requestMap); err != nil {
		return nil, fmt.Errorf("aigTokenRequestHook: unable to unmarshal request: %w", err)
	}

	if expireIn, ok := requestMap["expire_in"]; ok {
		// Remove expire_in if it's nil or if value/unit are zero/empty
		remove := false
		switch v := expireIn.(type) {
		case nil:
			remove = true
		case map[string]interface{}:
			unit, _ := v["unit"].(string)
			value, _ := v["value"].(float64)
			if unit == "" || value == 0 {
				remove = true
			}
		}
		if remove {
			delete(requestMap, "expire_in")
		}
	}

	modifiedBody, err := json.Marshal(requestMap)
	if err != nil {
		return nil, fmt.Errorf("aigTokenRequestHook: unable to marshal modified request: %w", err)
	}

	req.Body = io.NopCloser(strings.NewReader(string(modifiedBody)))
	req.ContentLength = int64(len(modifiedBody))

	return req, nil
}
