GO_SRCS := $(shell find ./go -name '*.go')
GO_BINS := $(GO_SRCS:%.go=%)
ZIG_SRCS := $(shell find ./zig -name '*.zig')
ZIG_BINS := $(ZIG_SRCS:%.zig=%)
SERVER := 'go'
PROBLEM := '0'

bins: go_bins zig_bins

go_bins: $(GO_BINS)

zig_bins: $(ZIG_BINS)

test: bins
	cd ./tester/$(PROBLEM) && SERVER_BIN='../../$(SERVER)/$(PROBLEM)' SERVER_ARGS='127.0.0.1:9999' go test -v

run: clean bins
	./$(SERVER)/$(PROBLEM) $(SERVER_ARGS)

$(GO_BINS): %: %.go
	go build -o $@ $<

$(ZIG_BINS): %: %.zig
	zig build-exe -femit-bin=$@ $<

clean:
	rm -f $(GO_BINS)
	rm -rf $(ZIG_BINS)
