SHELL=/bin/bash

COLOUR_NORMAL=$(shell tput sgr0)
COLOUR_RED=$(shell tput setaf 1)
COLOUR_GREEN=$(shell tput setaf 2)

json=false
rebuild=false
tail=false
start=false
imports=false

export OUT_DIR?=tmp
DIST_DIR?=./dist

COVERAGE?=90.0
# -- HELPERS ------------------------------------------------------------

.PHONY: all
all: clean vendor lint test docker-integration ## Perform clean, vendor, build, lint & test
	@if [[ -e .git/rebase-merge ]]; then git --no-pager log -1 --pretty='%h %s'; fi
	@printf '%sSuccess%s\n' "${COLOUR_GREEN}" "${COLOUR_NORMAL}"
	@echo "Run make help to view more commands"

.PHONY: clean
clean: ## Run go clean and remove generated binaries and coverage files
	go clean ./...
	rm -Rf ./${OUT_DIR} ${DIST_DIR}

.PHONY: lint
lint: imports fmt ## Runs the golangci-lint checker
	golangci-lint run

.PHONY: imports
imports:  ## Re-formats code using goimports
	goimports -v -w -e $(shell find . -type f -name '*.go' -not -path "*/vendor/*")

.PHONY: fmt
fmt: ## Run gofumpt on all the source code
	gofumpt -l -w -e $(shell find . -type f -name '*.go' -not -path "*/vendor/*")

.PHONY: generate
generate: ## Generate code from go:generate statements
	go generate -x -v ./...

.PHONY: vendor
vendor: download

download: ## Clean up go mod dependencies and download all dependencies.
	go mod tidy
	go mod download

# -- UNIT TESTING ------------------------------------------------------------
COVER_HTML_OUTPUT?=./${OUT_DIR}/coverage.html
COVER_JSON_OUTPUT?=./${OUT_DIR}/coverage.json
COVER_RAW_OUTPUT?=./${OUT_DIR}/coverage.out
COVEROUT?=./${OUT_DIR}/toolcover.out
COVER_TEXT_OUTPUT?=./${OUT_DIR}/coverage.txt
FLAGS=-covermode=atomic -coverprofile=${COVER_RAW_OUTPUT} -race
.PHONY: test
test: CC=
test: ## Test coverage of the currently checked out code using the locally installed version of Golang
	mkdir -p ./${OUT_DIR}
ifeq (${json}, true)
	set -o pipefail && go test ./... -json ${FLAGS} | tee $(COVER_JSON_OUTPUT)
else
	go test ./... ${FLAGS}
endif
	@go tool cover -func=$(COVER_RAW_OUTPUT) | tee $(COVEROUT)
	@go tool cover -func=$(COVER_RAW_OUTPUT) | $(CHECK_COVERAGE)
	@echo $

.PHONY: test-cover
test-cover: test
	@go tool cover -func=$(COVER_RAW_OUTPUT) | tee $(COVER_TEXT_OUTPUT)

.PHONY: test-cover-visual
test-cover-visual: test-cover ## Visual test coverage of the currently checked out code using the locally installed version of Golang
	@go tool cover -html=$(COVER_RAW_OUTPUT) -o $(COVER_HTML_OUTPUT)

# -- Docker Environment ------------------------------------------------
export APIC_CLIENT_ID?=THIS_IS_A_FAKE_CLIENT_ID
export AUDITLOG_PUBSUB_PORT?=8086
export BUILD_LOG_URL?=https://console.cloud.google.com/cloud-build/builds/build?project=project
export COMMIT_SHORT_SHA?=$(shell git rev-parse HEAD)
export TEST_CONFIG_FILE?=./config/ci.yaml
export PNV_TEST_CONFIG_FILE?=../../config/ci.yaml
export REPO_URL?=https://github.com/anzx/fabric-cards
export VERSION?=v0.0.0-local
export API_PORT?=8080
export OPS_PORT?=8082

export BASE_IMAGE?=gcr.io/anz-x-fabric-np-641432/platform/golang-1.18:v1.1.0
export BASE_CARDS_IMAGE?=base_cards:${COMMIT_SHORT_SHA}
export BASE_RUNTIME_IMAGE=gcr.io/anz-x-fabric-np-641432/base-images/base-debian11:v1.1.0

export CALLBACK_CONTAINER_TAG?=${CALLBACK_IMAGE}
export CALLBACK_IMAGE?=callback:${COMMIT_SHORT_SHA}
export CALLBACK_OPS_PORT?=8062
export CALLBACK_PORT?=8060

export CARDCONTROLS_CONTAINER_TAG?=${CARDCONTROLS_IMAGE}
export CARDCONTROLS_IMAGE?=cardcontrols:${COMMIT_SHORT_SHA}
export CARDCONTROLS_OPS_PORT?=8072
export CARDCONTROLS_PORT?=8070

export CARDS_CONTAINER_TAG?=${CARDS_IMAGE}
export CARDS_IMAGE?=cards:${COMMIT_SHORT_SHA}
export CARDS_OPS_PORT?=8082
export CARDS_PORT?=8080

export PNV_IMAGE?=pnv:${COMMIT_SHORT_SHA}
export INTEGRATION_IMAGE?=cards-integration:${COMMIT_SHORT_SHA}

export JAEGER_IMAGE?=hub.artifactory.gcp.anz/jaegertracing/all-in-one:1.30.0
export PROMETHEUS_IMAGE?=hub.artifactory.gcp.anz/prom/prometheus:v2.20.1
export PUBSUB_EMULATOR_IMAGE?=hub.artifactory.gcp.anz/jessejacksondocker/pubsub-emulator:293.0.0
export PUBSUB_PORT?=8185
export REDIS_EMULATOR_IMAGE?=hub.artifactory.gcp.anz/redis:5.0.13
export STUBS_GRPC_PORT?=9060
export STUBS_HTTP_PORT?=9070
export STUBS_IMAGE?=stubs:${COMMIT_SHORT_SHA}
export STUBS_LATENCY_ENABLED?=false

export VISA_GATEWAY_REGISTRY?=gcr.io/anz-x-fabric-np-641432/visa-gateway
export VISA_GATEWAY_VERSION?=v1.2.0-rc.30
export VISA_GATEWAY_API_PORT?=7080
export VISA_GATEWAY_IMAGE?=${VISA_GATEWAY_REGISTRY}/visa-gateway:${VISA_GATEWAY_VERSION}
export VISA_GATEWAY_OPS_PORT?=7082
export VISA_GATEWAY_STUB_GRPC_PORT?=7070
export VISA_GATEWAY_STUB_HTTP_PORT?=7060
export VISA_GATEWAY_STUB?=${VISA_GATEWAY_REGISTRY}/stub:${VISA_GATEWAY_VERSION}

export GSM_EMULATOR_HOST?=stubs:${STUBS_GRPC_PORT}
export GSM_EMULATOR_HOST_VISA?=http://visa-stub:${VISA_GATEWAY_STUB_HTTP_PORT}

# -- Docker ------------------------------------------------------------
.PHONY: docker-build-base
docker-build-base: ## Build base image for use in later builds.
	docker build -t ${BASE_CARDS_IMAGE} . --build-arg BASE_IMAGE=${BASE_IMAGE}

.PHONY: docker-build
docker-build: docker-build-base
	docker-compose build --parallel

.PHONY: docker-run
docker-run: docker-build docker-stubs ## Spin up all cards services
	docker-compose up -d cards cardcontrols callback

.PHONY: docker-cards
docker-cards: docker-build-base  ## Spin up cards
	docker-compose build cards --parallel
	docker-compose up --abort-on-container-exit cards

.PHONY: docker-stubs
docker-stubs: docker-support ## Run the stubs built in this repo
	docker-compose up -d stubs

.PHONY: docker-stop
docker-stop: ## Turn down the docker environment
	docker-compose down --remove-orphans

.PHONY: docker-clean
docker-clean: docker-stop ## Clean up your local docker system
	docker system prune

.PHONY: docker-integration
docker-integration: docker-build docker-integration-callback docker-integration-cardcontrols docker-integration-cards ## Start all services and run integration tests

.PHONY: docker-integration-callback
docker-integration-callback: ## Start Callback service and run integration tests
	docker-compose up --abort-on-container-exit integration-callback

.PHONY: docker-integration-cardcontrols
docker-integration-cardcontrols: ## Start Cardcontrols services and run integration tests
	docker-compose build --parallel integration-cardcontrols-grpc integration-cardcontrols-rest
	docker-compose up --abort-on-container-exit integration-cardcontrols-grpc
	docker-compose up --abort-on-container-exit integration-cardcontrols-rest

.PHONY: docker-integration-cards
docker-integration-cards: ## Start Cards services and run integration tests
	docker-compose build --parallel integration-cards-grpc integration-cards-rest
	docker-compose up --abort-on-container-exit integration-cards-grpc
	docker-compose up --abort-on-container-exit integration-cards-rest

.PHONY: docker-pnv
docker-pnv: docker-build docker-pnv-cards

.PHONY: docker-pnv-cards
docker-pnv-cards: docker-stop
	docker-compose build --no-cache --parallel pnv-cards
	docker-compose up pnv-cards


.PHONY: docker-support
docker-support: ## Start all supporting images (visa-gateway visa-stub pubsub auditlog redis)
	docker-compose up -d visa-gateway visa-stub pubsub auditlog redis

.PHONY: docker-logs
docker-logs: ## Read docker logs into the foreground
	docker-compose logs -f

.PHONY: docker-ops
docker-ops: ## Start prometheus and jaeger
	docker-compose up -d jaeger prometheus
	@echo "${COLOUR_GREEN}view prometheus ${COLOUR_BLUE}http://localhost:9090 ${COLOUR_NORMAL}"
	@echo "${COLOUR_GREEN}view jaeger ${COLOUR_BLUE}http://localhost:16686 ${COLOUR_NORMAL}"
	@open http://localhost:16686
	@open http://localhost:9090

# -- UTILS --------------------------------------------------------------

.DEFAULT_GOAL := help
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -d | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-25s\033[0m %s\n", $$1, $$2}'
	@echo "Coverage expected: ${COLOUR_GREEN}${COVERAGE}%${COLOUR_NORMAL}"

NAME ?= "CAMERON V"
.PHONY: make genjwt
genjwt:
	@go run test/script/genjwt.go --name="${NAME}"

define CHECK_COVERAGE
awk \
  -F '[ 	%]+' \
  -v threshold="$(COVERAGE)" \
  '/^total:/ { print; if ($$3 < threshold) { exit 1 } }' || { \
	printf '%sFAIL - Coverage below %s%%%s\n' \
	  "$(COLOUR_RED)" "$(COVERAGE)" "$(COLOUR_GREEN)"; \
	exit 1; \
  }
endef

CX=$(if $(shell which checkmarx), checkmarx, $(error No checkmarx in $$PATH, consider `go install github.com/anzx/fabric-entitlements/scripts/checkmarx@latest`))
GH=$(if $(shell which gh), gh, $(error No gh in $$PATH, consider `brew install gh`))

# Zip entire repo, excluding directories/files that don't need to be scanned
checkmarx-payload.zip:
	@mv -n .checkmarx-excludes .gitattributes
	@git archive $(shell git rev-parse HEAD) -o checkmarx-payload.zip --worktree-attributes
	@mv -n .gitattributes .checkmarx-excludes

CX_USERNAME ?= $(shell read -p "Checkmarx username: " pwd; echo $$pwd)
CX_PASSWORD ?= $(shell stty -echo; read -sp "Checkmarx password: 🔒" pwd; stty echo; echo $$pwd)
checkmarx-report.%: checkmarx-payload.zip
	@echo \*\*\*
	@CX_REPORT_FORMAT=$* \
	 CX_USERNAME=$(CX_USERNAME) \
	 CX_PASSWORD=$(CX_PASSWORD) \
	 CX_PROJECTID=498 \
	 COMMIT_SHA=$(shell git rev-parse HEAD) \
	 checkmarx

# Checks if this is a clean release tag (v\d\.\d\.\d), and publishes
# checkmarx-report.pdf to the tagged release using gh
GIT_TAG ?= $(shell git describe --tags)
.PHONY: publish-checkmarx-report
publish-checkmarx-report: clean-tag checkmarx-report.pdf checkmarx-report.csv
	$(GH) release upload $(GIT_TAG) checkmarx-report.pdf checkmarx-report.csv

BD=$(if $(shell which blackduck), blackduck, $(error No blackduck in $$PATH, consider `go install github.com/anzx/fabric-actions/releaserocket-cli/blackduck@latest`))
BD_API_TOKEN ?= $(shell read -p "Blackduck API Token: " pwd; echo $$pwd)
# Please run go install github.com/anzx/fabric-actions/releaserocket-cli/blackduck@latest
# For release work, ensure the tag has finished running its blackduck scan
# You will need to retrieve an API token from the blackduck ui before running
blackduck.json: clean-tag
	BD_PATH=https://blackduck.platform-blackduck.services.x.gcp.anz/ \
	BD_PROJECT=anzx/fabric-cards \
	BD_VERSION="Tag: $(GIT_TAG)" \
	BD_OUTPUT_NAME=$@ \
	BD_API_TOKEN=$(BD_API_TOKEN) \
	$(BD)

.PHONY: publish-blackduck-report
publish-blackduck-report: blackduck.json
	$(GH) release upload $(GIT_TAG) blackduck.json

.PHONY: publish-release-reports
publish-release-reports: publish-checkmarx-report publish-blackduck-report

clean-tag: # Check the current tag is a "clean" tag (example vx.y.z)
	[[ "$(GIT_TAG)" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$$ ]] || (echo "$(GIT_TAG) is not a clean release tag" && exit 1);

hooks: # Enable repo githooks on your local machine
	git config core.hooksPath .githooks

hotfix-branch: # Checkout the current latest tag as a hotfix branch (override GIT_TAG value to specify an previous tag)
	git checkout tags/$(GIT_TAG) -b hotfix/$(GIT_TAG)

