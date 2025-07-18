#!/bin/bash
set -e

# These variables are set from the repository's secrets and variables.
# Please contact a maintainer if these need to be updated.
export SCW_ACCESS_KEY=${E2E_SCW_ACCESS_KEY:?Variable not set or empty}
export SCW_SECRET_KEY=${E2E_SCW_SECRET_KEY:?Variable not set or empty}
export SCW_PROJECT_ID=${E2E_SCW_PROJECT_ID:?Variable not set or empty}

export SCW_REGION="nl-ams"
export CONTROL_PLANE_FAILURE_DOMAINS="[nl-ams-1, nl-ams-2, nl-ams-3]"
export CONTROL_PLANE_MACHINE_IMAGE="cluster-api-ubuntu-2404-v1.32.4"
export WORKER_MACHINE_IMAGE="cluster-api-ubuntu-2404-v1.32.4"
export KUBERNETES_VERSION="1.32.4"
export CONTROL_PLANE_MACHINE_COMMERCIAL_TYPE="PLAY2-NANO"
export WORKER_MACHINE_COMMERCIAL_TYPE="PLAY2-NANO"

make -C "$(dirname "$0")/../" test-e2e
