.PHONY: help submodules build smoke test

help:
	@echo "Targets: submodules | build | smoke | test"

submodules:
	git submodule update --init --recursive

build:
	mkdir -p bin
	go build -o bin/wisp .

# Compile CLI + reference standalone binary (green packaging path)
smoke: submodules build
	cd examples/reference-standalone && go mod tidy && go build -o reference-standalone .
	@echo ""
	@echo "OK: bin/wisp and examples/reference-standalone/reference-standalone"
	@echo "Run (needs wisp.yml with enabled connectors):"
	@echo "  ./examples/reference-standalone/reference-standalone \\"
	@echo "    --config ./examples/reference-standalone \\"
	@echo "    --wisp ./wisp.yml"

test:
	go test ./internal/services/live/manager/ ./internal/services/compile/ ./sdk/pkg/lifecycle/ ./sdk/pkg/runtime/ -count=1
