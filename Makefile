TARGET_DIR ?= ./dist
GOPROXY ?= https://goproxy.cn,direct
ARCH ?= amd64
OS ?= linux
APP ?= soc-event
BUILDX ?= false
PLATFORM ?= linux/amd64,linux/arm64
ORG ?= core.harbor1.domain/tool
TAG ?= v0.0.1

.PHONY: build vendor image push

all: image push

build:
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) GOPROXY=$(GOPROXY) go build -o $(TARGET_DIR)/$(ARCH)/$(APP) main.go

vendor:
	go mod vendor

image:
	DOCKER_BUILDKIT=1 docker build \
		--build-arg GOPROXY=$(GOPROXY) \
		--build-arg APP=$(APP) \
		--force-rm \
		--no-cache \
		-t $(ORG)/$(APP):$(TAG) \
		.
	
push:
	docker push $(ORG)/$(APP):$(TAG)