.PHONY: *

all: npa_publisher_tf_provider

docs:
	go generate ./...

terraform-provider-ns-npa-publisher: terraform-provider-ns-npa-publisher/npa_publisher.yaml
	speakeasy generate sdk --lang terraform -o terraform-provider-ns-npa-publisher -s terraform-provider-ns-npa-publisher/npa_publisher.yaml

terraform-provider-ns-npa-publisher/npa_publisher.yaml: check-speakeasy check-local-oas
	mkdir -p terraform-provider-ns-npa-publisher
	cp ${NETSKOPE_LOCAL_OAS_REPO}/endpoints/infrastructure/npa_publisher.yaml terraform-provider-ns-npa-publisher/npa_publisher.yaml
	speakeasy overlay apply -s ${NETSKOPE_LOCAL_OAS_REPO}/endpoints/infrastructure/npa_publisher.yaml -o ${NETSKOPE_LOCAL_OAS_REPO}/endpoints/infrastructure/npa_publisher_tf.yaml > terraform-provider-ns-npa-publisher/npa_publisher.yaml

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

