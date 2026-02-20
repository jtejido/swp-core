GO ?= go
PODMAN_COMPOSE ?= podman compose -f podman-compose.yml
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod
GOENV = GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE)
SPEC_VECTOR_ARGS ?=
CORE_PATTERN ?= conformance/vectors/core_*.json
# conformance-all intentionally relies on runner default discovery for official spec vectors.
CONFORMANCE_DIR ?= artifacts/conformance
CORE_DEFAULT_JSON ?= $(CONFORMANCE_DIR)/core.default.json
CORE_STRICT_JSON ?= $(CONFORMANCE_DIR)/core.strict.json
ALL_DEFAULT_JSON ?= $(CONFORMANCE_DIR)/all.default.json
ALL_STRICT_JSON ?= $(CONFORMANCE_DIR)/all.strict.json
CONFORMANCE_BUNDLE ?= $(CONFORMANCE_DIR)/swp-conformance-bundle.tar.gz

.PHONY: build test gen-vectors gen-c1-vectors gen-remaining-vectors vectors vectors-strict conformance-core conformance-all conformance-summary conformance-pack poc-vectors spec-vectors run-server run-client run-gateway demo clean clean-artifacts podman-up podman-down podman-logs podman-demo podman-poc-vectors podman-spec-vectors podman-vectors mcp-curl

build:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) build ./poc/cmd/swp-server ./poc/cmd/swp-client ./poc/cmd/vector-runner ./poc/cmd/spec-vector-runner ./poc/cmd/gen-vectors ./poc/cmd/gen-c1-vectors ./poc/cmd/gen-remaining-vectors ./poc/cmd/mcp-json-gateway

test:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) test ./poc/...

gen-vectors:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/gen-vectors

gen-c1-vectors:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/gen-c1-vectors

gen-remaining-vectors:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/gen-remaining-vectors

poc-vectors:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/vector-runner -pattern "conformance/vectors/poc_*.json"

vectors: spec-vectors

vectors-strict:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner -no-fallback $(SPEC_VECTOR_ARGS)

conformance-core:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(CONFORMANCE_DIR)
	@$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner -pattern "$(CORE_PATTERN)" -json-out "$(CORE_DEFAULT_JSON)" >"$(CONFORMANCE_DIR)/core.default.log" 2>&1
	@$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner -no-fallback -pattern "$(CORE_PATTERN)" -json-out "$(CORE_STRICT_JSON)" >"$(CONFORMANCE_DIR)/core.strict.log" 2>&1 || true
	@test -f "$(CORE_STRICT_JSON)" || echo '{"total":0,"passed":0,"failed":0,"fallback_count":0}' >"$(CORE_STRICT_JSON)"
	@python3 -c "import json,sys; p=sys.argv[1]; d=json.load(open(p)); print(f\"CORE default: total={d['total']} passed={d['passed']} failed={d['failed']} fallback={d.get('fallback_count', 0)} json={p}\")" "$(CORE_DEFAULT_JSON)"
	@python3 -c "import json,sys; p=sys.argv[1]; d=json.load(open(p)); print(f\"CORE strict: total={d['total']} passed={d['passed']} failed={d['failed']} fallback={d.get('fallback_count', 0)} json={p}\")" "$(CORE_STRICT_JSON)"

conformance-all:
	@mkdir -p $(GOCACHE) $(GOMODCACHE) $(CONFORMANCE_DIR)
	@$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner -json-out "$(ALL_DEFAULT_JSON)" >"$(CONFORMANCE_DIR)/all.default.log" 2>&1
	@$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner -no-fallback -json-out "$(ALL_STRICT_JSON)" >"$(CONFORMANCE_DIR)/all.strict.log" 2>&1 || true
	@test -f "$(ALL_STRICT_JSON)" || echo '{"total":0,"passed":0,"failed":0,"fallback_count":0}' >"$(ALL_STRICT_JSON)"
	@python3 -c "import json,sys; p=sys.argv[1]; d=json.load(open(p)); print(f\"ALL default: total={d['total']} passed={d['passed']} failed={d['failed']} fallback={d.get('fallback_count', 0)} json={p}\")" "$(ALL_DEFAULT_JSON)"
	@python3 -c "import json,sys; p=sys.argv[1]; d=json.load(open(p)); print(f\"ALL strict: total={d['total']} passed={d['passed']} failed={d['failed']} fallback={d.get('fallback_count', 0)} json={p}\")" "$(ALL_STRICT_JSON)"

conformance-summary:
	@$(MAKE) --no-print-directory conformance-core >/dev/null
	@python3 -c "import json,sys; d=json.load(open(sys.argv[1])); s=json.load(open(sys.argv[2])); sha=d.get('run', {}).get('runner_git_sha', 'nogit'); print(f\"SWP conformance (Core vectors) @ runner_git_sha={sha}\"); print(f\"CORE default: total={d['total']} passed={d['passed']} failed={d['failed']} fallback={d.get('fallback_count', 0)} json={sys.argv[1]}\"); print(f\"CORE strict: total={s['total']} passed={s['passed']} failed={s['failed']} fallback={s.get('fallback_count', 0)} json={sys.argv[2]}\")" "$(CORE_DEFAULT_JSON)" "$(CORE_STRICT_JSON)"

conformance-pack:
	@$(MAKE) --no-print-directory conformance-core >/dev/null
	@$(MAKE) --no-print-directory conformance-all >/dev/null
	@mkdir -p "$(CONFORMANCE_DIR)"
	@tar -czf "$(CONFORMANCE_BUNDLE)" \
		artifacts/conformance/*.json \
		artifacts/conformance/*.log \
		docs/conformance.md \
		docs/spec-vector-runner-output.md \
		docs/error-codes.md \
		docs/publication-artifact-index.md
	@echo "bundle: $(CONFORMANCE_BUNDLE)"

spec-vectors:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/spec-vector-runner $(SPEC_VECTOR_ARGS)

run-server:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/swp-server -listen :7777

run-client:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/swp-client -addr 127.0.0.1:7777

run-gateway:
	mkdir -p $(GOCACHE) $(GOMODCACHE)
	$(GOENV) $(GO) run ./poc/cmd/mcp-json-gateway -listen :8080 -swp 127.0.0.1:7777

demo:
	@set -e; \
	mkdir -p $(GOCACHE) $(GOMODCACHE); \
	$(GOENV) $(GO) run ./poc/cmd/swp-server -listen :7777 >/tmp/swp-server.log 2>&1 & \
	SERVER_PID=$$!; \
	trap 'kill $$SERVER_PID 2>/dev/null || true' EXIT; \
	sleep 1; \
	$(GOENV) $(GO) run ./poc/cmd/swp-client -addr 127.0.0.1:7777

mcp-curl:
	curl -sS -X POST http://127.0.0.1:8080/mcp \
	  -H 'content-type: application/json' \
	  -d '{"jsonrpc":"2.0","id":"demo-1","method":"tools/list","params":{}}' | jq .

podman-up:
	$(PODMAN_COMPOSE) up -d --build swp-server mcp-json-gateway

podman-demo:
	$(PODMAN_COMPOSE) up --build swp-server swp-client

podman-poc-vectors:
	$(PODMAN_COMPOSE) run --rm --build swp-vectors

podman-spec-vectors:
	$(PODMAN_COMPOSE) run --rm --build swp-spec-vectors

podman-vectors: podman-spec-vectors

podman-logs:
	$(PODMAN_COMPOSE) logs -f swp-server mcp-json-gateway

podman-down:
	$(PODMAN_COMPOSE) down --remove-orphans

clean:
	rm -f swp-server swp-client vector-runner spec-vector-runner gen-vectors mcp-json-gateway
	rm -rf .cache

clean-artifacts:
	rm -rf artifacts/conformance
