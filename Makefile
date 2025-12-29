# The binary to build (just the basename).
BIN ?= velero-plugin-for-jdcloud

# Where to push the docker image.
REGISTRY ?= velero

# Image name
IMAGE ?= $(REGISTRY)/$(BIN)

# Image tag
VERSION ?= latest

# Build arch
ARCH ?= amd64

# Target OS
TargetOS ?= linux

# Build the binary
.PHONY: build
build:
	go build -v -o $(BIN) ./velero-plugin-for-jdcloud

# Build the docker image
.PHONY: docker-build
docker-build:
	docker build -t $(IMAGE):$(VERSION) --build-arg TARGETOS=$(TargetOS) --build-arg TARGETARCH=$(ARCH) .

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push $(IMAGE):$(VERSION)

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(BIN)

# Run tests
.PHONY: test
test:
	go test -v ./...