.PHONY: *

all: terraform-provider-ns-npa-publisher

docs:
	go generate ./...

terraform-provider-ns-npa-publisher: netskope.yaml
	speakeasy generate sdk --lang terraform -o . -s netskope.yaml
	#go run ../../cmd/generate/main.go -l terraform -o . -s netskope.yaml


netskope.yaml: check-speakeasy check-local-oas
	speakeasy overlay apply -s ${NETSKOPE_LOCAL_OAS_REPO}/base_oas.yaml -o ${NETSKOPE_LOCAL_OAS_REPO}/terraform_overlay.yaml > netskope.yaml

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

