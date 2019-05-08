ifndef NONINTERACTIVE
include ../make/help.mk
include ../make/config.mk
include ../make/toolchain.mk
endif

VAULT_CLIENT_PREFIX = gcr.io/${GCLOUD_PROJECT_ID}/vault-client-alpine
VAULT_CLIENT_TAGGED_CONTAINER_NAME = "${VAULT_CLIENT_PREFIX}:${DEPLOY_TAG}"
VAULT_CLIENT_LATEST_CONTAINER_NAME = "${VAULT_CLIENT_PREFIX}:latest"

image:
	docker build -t ${VAULT_CLIENT_PREFIX}:dex-local .