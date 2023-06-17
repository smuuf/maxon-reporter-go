MAIN_ENTRYPOINT_FILE=./main.go
BINARY_NAME=reporter
BINARY_DIR=build
BINARY_PATH=$(BINARY_DIR)/$(BINARY_NAME)
VERSION=`date +%Y.%m.%d.%H%M`
DEV_FLAG_NAME="-dev"

# Unless explicily requested by calling "make build-release" all builds are
# "dev" builds by default.
BUILD_DEV?=1

LDFLAGS=-ldflags "-w -s \
-X 'reporter/internal.ReporterVersion=$(VERSION)' \
-X 'reporter/internal.ReporterDevFlag=$(BUILD_DEV)' \
"

ifneq "$(BUILD_DEV)" "0"
	VERSION:=$(VERSION)$(DEV_FLAG_NAME)
endif

.PHONY: build build-release run clean foreground try test

build:
	mkdir -p $(BINARY_DIR)
	go build -o ./$(BINARY_PATH) $(LDFLAGS) $(MAIN_ENTRYPOINT_FILE)

build-release:
	# Request non-dev-build.
	BUILD_DEV=0 $(MAKE) build
	# Pack the built binary via upx (we get smaller binary).
	upx-ucl ./$(BINARY_PATH)

run: build
	./$(BINARY_PATH)

try: build
	./$(BINARY_PATH) --try

foreground: build
	./$(BINARY_PATH) --foreground

test:
	go test -cover ./...
	go test -bench=. ./...

clean:
	go clean
	rm ./$(BINARY_NAME)
