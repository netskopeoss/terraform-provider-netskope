package provider

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	sdkerrors "github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/errors"
)

// debugResponse deliberately reports only allowlisted failure metadata. Normal
// Terraform diagnostics must not include HTTP bodies, URLs, or other headers.
// It neither reads the body nor mutates the response or its request.
func debugResponse(response *http.Response) string {
	method, status, requestID := "unavailable", "unavailable", "unavailable"
	if response != nil {
		status = fmt.Sprintf("%d", response.StatusCode)
		if response.Request != nil && response.Request.Method != "" {
			method = response.Request.Method
		}
		if value := response.Header.Get("X-Netskope-Request-Id"); value != "" {
			requestID = value
		}
	}
	return formatAPIFailure(method, status, requestID)
}

// SDK error strings can contain complete response bodies; transport error
// strings can contain URLs and credentials. Never pass either through to a
// normal Terraform diagnostic.
func apiErrorDetails(err error) string {
	var failure *apiFailureError
	if errors.As(err, &failure) && failure != nil {
		return debugResponse(failure.response)
	}
	var sdkErr *sdkerrors.SDKError
	if errors.As(err, &sdkErr) && sdkErr != nil {
		if sdkErr.RawResponse != nil {
			return debugResponse(sdkErr.RawResponse)
		}
		return formatAPIFailure("unavailable", fmt.Sprintf("%d", sdkErr.StatusCode), "unavailable")
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr != nil {
		return formatAPIFailure(strings.ToUpper(urlErr.Op), "unavailable", "unavailable")
	}
	return debugResponse(nil)
}

// Preserve metadata from hand-written HTTP operations without retaining bodies.
type apiFailureError struct{ response *http.Response }

func (e *apiFailureError) Error() string { return debugResponse(e.response) }

func formatAPIFailure(method, status, requestID string) string {
	// Quote externally supplied strings so control characters cannot inject lines.
	return fmt.Sprintf("method=%q status=%s request_id=%q", method, status, requestID)
}
