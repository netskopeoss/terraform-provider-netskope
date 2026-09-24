package hooks

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// npaTemplatesAPIEntry represents one element from /api/v2/templates/usernotifications.
type npaTemplatesAPIEntry struct {
	FileName   string `json:"file_name"`
	Name       string `json:"name"`
	ActionType string `json:"action_type"`
}

// npaTemplatesAPIResponseBody is the response envelope from /api/v2/templates/usernotifications.
type npaTemplatesAPIResponseBody struct {
	TotalCount int                    `json:"total_count"`
	Elements   []npaTemplatesAPIEntry `json:"elements"`
}

// npaTemplatesAPICache is a session-level cache populated from /api/v2/templates/usernotifications.
// It provides bidirectional name↔filename lookup scoped by action_type, enabling users to specify
// templates by display name for any action_name (including periodic_reauth).
//
// The cache is populated lazily on the first rule operation that involves a template field.
// If the endpoint is unavailable on the tenant, the cache remains empty and the provider
// falls back to the session-based npaTemplateCache for block rules; periodic_reauth rules
// fall back to requiring the filename directly.
//
// Keys are "action_type:name" and "action_type:file_name" to avoid cross-action collisions
// (e.g. both "block" and "periodic_reauth" have a template named "Default Template" on some tenants).
//
// See docs/bugs/BUG-020-periodic-reauth-template.md
var npaTemplatesAPICache = &struct {
	mu     sync.Mutex
	loaded bool
	byFile map[string]string // "action_type:file_name" → display_name
	byName map[string]string // "action_type:display_name" → file_name
}{
	byFile: make(map[string]string),
	byName: make(map[string]string),
}

// ensureNPATemplatesAPILoaded fetches /api/v2/templates/usernotifications exactly once per
// session. Subsequent calls are no-ops. If the call fails or the endpoint is unavailable,
// the cache is left empty (fallback behaviour applies).
func ensureNPATemplatesAPILoaded(apiBase, token string) {
	npaTemplatesAPICache.mu.Lock()
	defer npaTemplatesAPICache.mu.Unlock()

	if npaTemplatesAPICache.loaded {
		return
	}
	// Mark loaded immediately so concurrent goroutines don't duplicate the call.
	// If the HTTP call fails, the cache stays empty and fallback behaviour applies.
	npaTemplatesAPICache.loaded = true

	if apiBase == "" || token == "" {
		log.Printf("INFO: npaTemplatesAPI: no base URL or token available, skipping load")
		return
	}

	url := apiBase + "/templates/usernotifications"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("WARN: npaTemplatesAPI: could not build request for %s: %v", url, err)
		return
	}
	req.Header.Set("Netskope-Api-Token", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("WARN: npaTemplatesAPI: request to %s failed: %v — template name resolution unavailable", url, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("WARN: npaTemplatesAPI: could not read response body: %v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("WARN: npaTemplatesAPI: endpoint returned %d — template name resolution unavailable", resp.StatusCode)
		return
	}

	// Guard against API-level errors returned as 200 OK with {"status":"error",...}.
	var errCheck struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &errCheck) == nil && errCheck.Status == "error" {
		log.Printf("WARN: npaTemplatesAPI: API returned error status — template name resolution unavailable")
		return
	}

	var parsed npaTemplatesAPIResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		log.Printf("WARN: npaTemplatesAPI: could not parse response: %v", err)
		return
	}

	for _, e := range parsed.Elements {
		npaTemplatesAPICache.byFile[e.ActionType+":"+e.FileName] = e.Name
		npaTemplatesAPICache.byName[e.ActionType+":"+e.Name] = e.FileName
	}
	log.Printf("INFO: npaTemplatesAPI: loaded %d templates", len(parsed.Elements))
}

// npaTemplatesAPIDisplayName returns the display name for a given action type + filename.
// Returns ("", false) if the cache is empty or the entry is not found.
// Falls back to a global (action_type-agnostic) search when the specific key is not
// found: template filenames are tenant-unique so the fallback is unambiguous.
func npaTemplatesAPIDisplayName(actionType, fileName string) (string, bool) {
	npaTemplatesAPICache.mu.Lock()
	defer npaTemplatesAPICache.mu.Unlock()
	if v, ok := npaTemplatesAPICache.byFile[actionType+":"+fileName]; ok {
		return v, true
	}
	// Fallback: search all action types — filenames are unique per tenant.
	// Necessary when the API returns a filename for a rule whose action_type
	// differs from the action_type recorded in the templates API metadata
	// (e.g. a "block"-typed template used in a periodic_reauth rule).
	suffix := ":" + fileName
	for k, v := range npaTemplatesAPICache.byFile {
		if strings.HasSuffix(k, suffix) {
			return v, true
		}
	}
	return "", false
}

// npaTemplatesAPIFileName returns the .html filename for a given action type + display name.
// Returns ("", false) if the cache is empty or the entry is not found.
// Falls back to a global (action_type-agnostic) search — display names are tenant-unique.
// Used by BeforeRequest to translate user-supplied display names to filenames before
// sending create/update payloads to the API (required for periodic_reauth rules, where
// the API stores the value verbatim rather than translating it server-side).
func npaTemplatesAPIFileName(actionType, displayName string) (string, bool) {
	npaTemplatesAPICache.mu.Lock()
	defer npaTemplatesAPICache.mu.Unlock()
	if v, ok := npaTemplatesAPICache.byName[actionType+":"+displayName]; ok {
		return v, true
	}
	// Fallback: search all action types.
	suffix := ":" + displayName
	for k, v := range npaTemplatesAPICache.byName {
		if strings.HasSuffix(k, suffix) {
			return v, true
		}
	}
	return "", false
}

// npaTemplatesAPISeedForTest directly populates the cache for unit tests without making
// a real HTTP call. Reset with npaTemplatesAPIResetForTest between tests.
func npaTemplatesAPISeedForTest(entries []npaTemplatesAPIEntry) {
	npaTemplatesAPICache.mu.Lock()
	defer npaTemplatesAPICache.mu.Unlock()
	npaTemplatesAPICache.loaded = true
	for _, e := range entries {
		npaTemplatesAPICache.byFile[e.ActionType+":"+e.FileName] = e.Name
		npaTemplatesAPICache.byName[e.ActionType+":"+e.Name] = e.FileName
	}
}

// npaTemplatesAPIResetForTest clears the cache and resets the loaded flag.
// Call this in TestMain or at the start of tests that need a clean cache.
func npaTemplatesAPIResetForTest() {
	npaTemplatesAPICache.mu.Lock()
	defer npaTemplatesAPICache.mu.Unlock()
	npaTemplatesAPICache.loaded = false
	npaTemplatesAPICache.byFile = make(map[string]string)
	npaTemplatesAPICache.byName = make(map[string]string)
}

// extractAPIBase extracts the API base URL (scheme://host/api/v2) from a request URL.
// The provider always uses /api/v2 as the root per CLAUDE.md, so this reliably finds
// the prefix for constructing sibling endpoint URLs.
// Returns "" if the URL is nil or does not contain /api/v2.
func extractAPIBase(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	path := req.URL.Path
	idx := strings.Index(path, "/api/v2")
	if idx < 0 {
		return ""
	}
	return req.URL.Scheme + "://" + req.URL.Host + path[:idx+len("/api/v2")]
}