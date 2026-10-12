GOTOOLCHAIN ?= go1.26.8
export GOTOOLCHAIN
MAIN := ./cmd/tinypod
HARNESS := nimbleterminal.dev/tinybubbles/web
HARNESS_DIR := $(shell go mod download $(HARNESS) && go list -m -f '{{.Dir}}' $(HARNESS))
include $(HARNESS_DIR)/application.mk
