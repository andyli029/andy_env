all: build

BIN=shard

run: build
	./${BIN}

build: shardtable
shardtable: clean
	rm -rf $(BIN)
	go build -gcflags "-N -l" -o ./$(BIN) ./shardtable.go ./config.go
clean:
	@rm -rf ./$(BIN)

test:
	echo "now is null"
