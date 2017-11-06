ROOT= $(shell echo $(GOPATH) | awk -F':' '{print $$1}')
PROJ_DIR= $(ROOT)/src
PWD = $(shell pwd)

all: build

build: kingshard
goyacc:
	go build -o ./bin/goyacc ./vendor/golang.org/x/tools/cmd/goyacc
kingshard: goyacc
	- unlink $(PROJ_DIR)/kingshard
	mkdir -p $(PROJ_DIR) && ln -s $(PWD) $(PROJ_DIR)/kingshard
	./bin/goyacc -o ./sqlparser/sql.go ./sqlparser/sql.y
	gofmt -w ./sqlparser/sql.go
	@bash genver.sh
	go build -o ./bin/kingshard ./cmd/kingshard
clean:
	@rm -rf bin
	@rm -f ./sqlparser/y.output ./sqlparser/sql.go

test:
	cd $(PROJ_DIR)/kingshard && go list ./... | grep -v vendor | xargs go test -v
