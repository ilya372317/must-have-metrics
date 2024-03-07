lint_dir = ./my-lint
lint_result_file = result.txt
lint_exec = mylint

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(shell pwd)/golangci-lint/cache:/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.55.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint

.PHONY: my-lint
my-lint: _run-lint _clear_lint_binary

.PHONY: _clear_lint_binary
_clear_lint_binary:
	rm $(lint_exec)

.PHONY: _run-lint
_run-lint: _create-lint-dir _build_linter
	-./$(lint_exec) ./... 2> $(lint_dir)/$(lint_result_file)

.PHONY: _build_linter
_build_linter:
	go build -o $(lint_exec) cmd/staticlint/main.go

.PHONY: _create-lint-dir
_create-lint-dir:
	mkdir -p $(lint_dir)

.PHONY: clear-my-lint
clear-my-lint:
	rm -rf $(lint_dir)