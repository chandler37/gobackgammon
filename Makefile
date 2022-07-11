.PHONY: help
help:
	@echo make fmt runs go fmt. Read the Makefile for the rest.

.PHONY: check
check:
	go test ./...

.PHONY: cover
cover:
	go test -coverprofile coverprofile.out ./... && go tool cover -html=coverprofile.out

.PHONY: bench
bench:
	go test -run ZZZ -bench=. ./...

.PHONY: benchseeded
benchseeded:
	go test -run ZZZ -bench=Reproduc ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: build
build:
	go build .

gobackgammon: bg.go ai/*.go brd/*.go json/*.go svg/*.go
	go build .

.PHONY: run
run: gobackgammon
	./gobackgammon -debug

.PHONY: run10
run10: gobackgammon
	./gobackgammon -debug -auto -match 10

.PHONY: runauto
runauto: gobackgammon
	./gobackgammon -auto

# TODO(chandler37): Before using go 1.11 modules (see go.mod), this worked. Fix it.
.PHONY: doc
doc:
	godoc -http=:6060

.PHONY: textdoc
textdoc:
	go doc github.com/chandler37/gobackgammon/ai
	@echo " "
	go doc github.com/chandler37/gobackgammon/brd
	@echo " "
	go doc github.com/chandler37/gobackgammon/json
	@echo " "
	go doc github.com/chandler37/gobackgammon/svg

.PHONY: clean
clean:
	rm -fr bin/ pkg/ coverprofile.out
	go clean -cache
	go clean ./...
