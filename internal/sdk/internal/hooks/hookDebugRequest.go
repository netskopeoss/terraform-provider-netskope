package hooks

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

// debugRequestHook logs all outgoing HTTP requests for debugging
type debugRequestHook struct{}

var (
	_ beforeRequestHook = (*debugRequestHook)(nil)
)

func (d *debugRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	// Write to a file since terraform captures stderr
	f, err := os.OpenFile("/tmp/netskope_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		fmt.Fprintf(f, "[DEBUG] ====== HTTP REQUEST ======\n")
		fmt.Fprintf(f, "[DEBUG] Operation: %s\n", hookCtx.OperationID)
		fmt.Fprintf(f, "[DEBUG] Method: %s\n", req.Method)
		fmt.Fprintf(f, "[DEBUG] URL: %s\n", req.URL.String())
		fmt.Fprintf(f, "[DEBUG] Headers:\n")
		for k, v := range req.Header {
			// Don't log the actual API token value
			if k == "Netskope-Api-Token" {
				fmt.Fprintf(f, "[DEBUG]   %s: [REDACTED]\n", k)
			} else {
				fmt.Fprintf(f, "[DEBUG]   %s: %v\n", k, v)
			}
		}

		// Log body if present
		if req.Body != nil {
			body, err := io.ReadAll(req.Body)
			if err == nil {
				fmt.Fprintf(f, "[DEBUG] Body: %s\n", string(body))
				// Restore the body for the actual request
				req.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
		fmt.Fprintf(f, "[DEBUG] =============================\n")
	}

	return req, nil
}
