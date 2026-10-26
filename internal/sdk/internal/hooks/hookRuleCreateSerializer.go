package hooks

import (
	"net/http"
	"sync"
)

// ruleCreateSerializer serializes concurrent create requests for NPA rules and
// NPA private apps to prevent backend race conditions.
//
// Rules (#66): the API does not use atomic sequences for rule_id generation. When
// concurrent POST requests hit /policy/npa/rules, both transactions read the same
// max ID and attempt to insert max+1, causing a MySQL duplicate primary key error.
//
// Private apps (#120): when two apps created in the same apply carry the same NEW
// tag_name, the API creates the tag row on first use; concurrent POSTs to
// /steering/apps/private both try to insert it and the second fails with
// "Duplicate entry '<tag>' for key 'tag_unique'". Serialising the creates means the
// second request finds the tag already present and attaches it by name.
//
// This hook acquires a mutex in BeforeRequest and releases it in AfterSuccess
// or AfterError, ensuring only one such create is in-flight at a time. Rules and
// private apps share one mutex; the cost is negligible relative to API latency.
//
// See docs/bugs/BUG-008-rule-creation-race-condition.md
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/66
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/120
type ruleCreateSerializer struct {
	mu sync.Mutex
}

// serializedCreateOps are the operations whose creates are serialised.
var serializedCreateOps = map[string]bool{
	"createNPARules":       true,
	"createNPAPrivateApps": true,
}

var (
	_ beforeRequestHook = (*ruleCreateSerializer)(nil)
	_ afterSuccessHook  = (*ruleCreateSerializer)(nil)
	_ afterErrorHook    = (*ruleCreateSerializer)(nil)
)

func (h *ruleCreateSerializer) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if serializedCreateOps[hookCtx.OperationID] {
		h.mu.Lock()
	}
	return req, nil
}

func (h *ruleCreateSerializer) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if serializedCreateOps[hookCtx.OperationID] {
		h.mu.Unlock()
	}
	return res, nil
}

func (h *ruleCreateSerializer) AfterError(hookCtx AfterErrorContext, res *http.Response, err error) (*http.Response, error) {
	if serializedCreateOps[hookCtx.OperationID] {
		h.mu.Unlock()
	}
	return res, err
}
