package hooks

import (
	"net/http"
	"sync"
)

// ruleCreateSerializer serializes concurrent rule creation requests to prevent
// a backend race condition where the API assigns duplicate rule IDs.
//
// The API does not use atomic sequences for rule_id generation. When concurrent
// POST requests hit /policy/npa/rules, both transactions read the same max ID
// and attempt to insert max+1, causing a MySQL duplicate primary key error.
//
// This hook acquires a mutex in BeforeRequest and releases it in AfterSuccess
// or AfterError, ensuring only one rule creation is in-flight at a time.
//
// See docs/bugs/BUG-008-rule-creation-race-condition.md
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/66
type ruleCreateSerializer struct {
	mu sync.Mutex
}

var (
	_ beforeRequestHook = (*ruleCreateSerializer)(nil)
	_ afterSuccessHook  = (*ruleCreateSerializer)(nil)
	_ afterErrorHook    = (*ruleCreateSerializer)(nil)
)

func (h *ruleCreateSerializer) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID == "createNPARules" {
		h.mu.Lock()
	}
	return req, nil
}

func (h *ruleCreateSerializer) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID == "createNPARules" {
		h.mu.Unlock()
	}
	return res, nil
}

func (h *ruleCreateSerializer) AfterError(hookCtx AfterErrorContext, res *http.Response, err error) (*http.Response, error) {
	if hookCtx.OperationID == "createNPARules" {
		h.mu.Unlock()
	}
	return res, err
}