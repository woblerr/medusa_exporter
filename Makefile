SHELL := /bin/bash
APP_NAME := medusa_exporter
BRANCH_FULL := $(shell git rev-parse --abbrev-ref HEAD)
BRANCH := $(subst /,-,$(BRANCH_FULL))
GIT_REV := $(shell git describe --abbrev=7 --always)
BUILD_DATE := $(shell date +%Y-%m-%dT%H:%M:%S%z)
BUILD_USER ?= medusa_exporter
SERVICE_CONF_DIR := /etc/systemd/system
HTTP_PORT := 19500
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
DOCKER_CONTAINER_E2E := $(shell docker ps -a -q -f name=$(APP_NAME)_e2e)
HTTP_PORT_E2E := $(shell echo $$((10000 + ($$RANDOM % 10000))))
LDFLAGS = -X github.com/prometheus/common/version.Version=$(BRANCH)-$(GIT_REV) \
		  -X github.com/prometheus/common/version.Branch=$(BRANCH) \
		  -X github.com/prometheus/common/version.Revision=$(GIT_REV) \
		  -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE) \
		  -X github.com/prometheus/common/version.BuildUser=$(BUILD_USER)

.PHONY: test
test:
	@echo "Run tests for $(APP_NAME)"
	TZ="Etc/UTC" go test -mod=vendor -timeout=60s -count 1  ./...

.PHONY: build
build:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
		-mod=vendor -trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(APP_NAME) $(APP_NAME).go

.PHONY: build-arm
build-arm:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
		-mod=vendor -trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(APP_NAME) $(APP_NAME).go

.PHONY: build-darwin
build-darwin:
	@echo "Build $(APP_NAME)"
	@make test
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build \
		-mod=vendor -trimpath \
		-ldflags "$(LDFLAGS)" \
		-o $(APP_NAME) $(APP_NAME).go

.PHONY: dist
dist:
	- @mkdir -p dist
	DOCKER_BUILDKIT=1 docker build -f Dockerfile.artifacts --progress=plain -t medusa_exporter_dist .
	- @docker rm -f medusa_exporter_dist 2>/dev/null || exit 0
	docker run -d --name=medusa_exporter_dist medusa_exporter_dist
	docker cp medusa_exporter_dist:/artifacts dist/
	docker rm -f medusa_exporter_dist


.PHONY: prepare-service
prepare-service:
	@echo "Prepare config file $(APP_NAME).service for systemd"
	cp $(ROOT_DIR)/$(APP_NAME).service.template $(ROOT_DIR)/$(APP_NAME).service
	sed -i.bak "s|/usr/bin|$(ROOT_DIR)|g" $(APP_NAME).service
	rm $(APP_NAME).service.bak

.PHONY: install-service
install-service:
	@echo "Install $(APP_NAME) as systemd service"
	$(call service-install)

.PHONY: remove-service
remove-service:
	@echo "Delete $(APP_NAME) systemd service"
	$(call service-remove)

define service-install
	cp $(ROOT_DIR)/$(APP_NAME).service $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl enable $(APP_NAME)
	systemctl restart $(APP_NAME)
	systemctl status $(APP_NAME)
endef

define service-remove
	systemctl stop $(APP_NAME)
	systemctl disable $(APP_NAME)
	rm $(SERVICE_CONF_DIR)/$(APP_NAME).service
	systemctl daemon-reload
	systemctl reset-failed
endef