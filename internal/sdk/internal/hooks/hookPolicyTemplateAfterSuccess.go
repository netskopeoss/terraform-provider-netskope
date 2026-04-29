package hooks

import (
	"net/http"
)

// policyTemplateResponseHook is a placeholder for future template name resolution.
// Currently, template drift is handled by the plan modifier in
// nparules_resource_planmodify.go (suppressTemplateDrift).
//
// If the notifications API (/api/v2/templates/usernotifications) becomes
// accessible to API tokens in the future, this hook can be activated to
// map file names back to display names in the GET response.
//
// See: https://github.com/netskopeoss/terraform-provider-netskope/issues/79
type policyTemplateResponseHook struct{}

var _ afterSuccessHook = (*policyTemplateResponseHook)(nil)

func (h *policyTemplateResponseHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	return res, nil
}
