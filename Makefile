.DEFAULT_GOAL := binary

binary:
	go build .

test:
	go test ./... -cover

bench:
	@go test -bench=. -benchtime=1ns ./segments/pass | grep "cpu:"
	@echo "results:"
	@go test -bench=. -run=Bench ./... | grep -E "^Bench" | awk '{fps = 1/(($$3)/1e9); sub(/Benchmark/, "", $$1); sub(/-.*/, "", $$1); printf("%15s: %8s ns/flow, %7.0f flows/s\n", tolower($$1), $$3, fps)}'
	@rm segments/output/sqlite/bench.sqlite

pb/flow.pb.go: pb/flow.proto
	protoc --go_out=. --go_opt=paths=source_relative pb/flow.proto
