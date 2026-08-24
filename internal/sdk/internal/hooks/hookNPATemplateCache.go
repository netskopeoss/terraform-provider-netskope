package hooks

import (
	"context"
	"strings"
	"sync"
)

// npaTemplateCtxKey is the context key used to pass a notification template
// display name from BeforeRequest to AfterSuccess for createNPARules operations.
type npaTemplateCtxKey struct{}

// npaTemplateCache maps NPA rule notification template file names (e.g. "2.html")
// to their display names (e.g. "tf-test-template"). The Netskope API accepts
// display names on create/update but returns .html file names on GET responses.
//
// The cache is populated by the AfterSuccess hook for createNPARules: BeforeRequest
// stores the display name in the request context, and AfterSuccess observes the
// file name in the response, stores the mapping, then substitutes the display name
// back into the response body. Subsequent getNPARules/updateNPARules responses are
// also fixed using the cache, so Terraform state always holds the display name.
//
// This eliminates the perpetual display-name/file-name drift for both block and
// periodic_reauth rules without requiring access to the notifications API.
// See docs/bugs/BUG-019-block-rule-template-phantom-update.md
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/116
var npaTemplateCache = &struct {
	mu sync.RWMutex
	m  map[string]string // ".html file name" → "display name"
}{m: make(map[string]string)}

func npaTemplateCacheGet(fileName string) (string, bool) {
	npaTemplateCache.mu.RLock()
	defer npaTemplateCache.mu.RUnlock()
	v, ok := npaTemplateCache.m[fileName]
	return v, ok
}

func npaTemplateCacheSet(fileName, displayName string) {
	if fileName == "" || displayName == "" || !strings.HasSuffix(fileName, ".html") {
		return
	}
	npaTemplateCache.mu.Lock()
	defer npaTemplateCache.mu.Unlock()
	npaTemplateCache.m[fileName] = displayName
}

// withNPATemplateDisplayName returns a new context carrying the given template
// display name, to be retrieved by npaTemplateDisplayNameFromCtx in AfterSuccess.
func withNPATemplateDisplayName(ctx context.Context, displayName string) context.Context {
	return context.WithValue(ctx, npaTemplateCtxKey{}, displayName)
}

// npaTemplateDisplayNameFromCtx retrieves the template display name stored by
// withNPATemplateDisplayName. Returns ("", false) if not present.
func npaTemplateDisplayNameFromCtx(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(npaTemplateCtxKey{}).(string)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}
