.PHONY: all docs terraform-provider-ns-npa-publisher check-local-oas

all: terraform-provider-ns-npa-publisher

docs:
	go generate ./...

terraform-provider-ns-npa-publisher: netskope.yaml
	speakeasy generate sdk --lang terraform -o . -s netskope.yaml
	#go run ../../cmd/generate/main.go -l terraform -o . -s netskope.yaml

spec/merged.yaml: check-speakeasy check-local-oas
	speakeasy merge -s "${NETSKOPE_LOCAL_OAS_REPO}/base_oas.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/steering/npa_apps_private.yaml" \
		-s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/infrastructure/npa_publisher.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/infrastructure/npa_upgrade_profiles.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/policy/npa_policy.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/policy/npa_policygroup.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/steering/gre.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/steering/ipsec.yaml" \
        -s "${NETSKOPE_LOCAL_OAS_REPO}/endpoints/steering/npa_private_tag.yaml" \
		-o spec/merged.yaml

netskope.yaml: check-speakeasy check-local-oas spec/merged.yaml
	cp spec/merged.yaml netskope.yaml
	#speakeasy overlay apply -s spec/merged.yaml -o ${NETSKOPE_LOCAL_OAS_REPO}/terraform_overlay.yaml > spec/netskope.yaml

check-speakeasy:
	@command -v speakeasy >/dev/null 2>&1 || { echo >&2 "speakeasy CLI is not installed. Please install before continuing."; exit 1; }

check-local-oas:
ifndef NETSKOPE_LOCAL_OAS_REPO
	$(error Environment variable NETSKOPE_LOCAL_OAS_REPO is undefined)
endif
ifneq ($(wildcard ${NETSKOPE_LOCAL_OAS_REPO}/.*),)
	@echo "Found ${NETSKOPE_LOCAL_OAS_REPO}."
else
	@echo "Did not find ${NETSKOPE_LOCAL_OAS_REPO}."
endif

