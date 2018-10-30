.PHONY: fmt help build clean check bench cover runauto vet doc textdoc run10 run
help:
	@echo make fmt runs go fmt. Read the Makefile for the rest.
check:
	go test ./...
cover:
	go test -coverprofile coverprofile.out ./... && go tool cover -html=coverprofile.out
bench:
	go test -run ZZZ -bench=. ./...
benchseeded:
	go test -run ZZZ -bench=Reproduc ./...
fmt:
	go fmt ./...
vet:
	go vet ./...
build:
	go build .
gobackgammon: bg.go ai/*.go brd/*.go json/*.go svg/*.go
	go build .
run: gobackgammon
	./gobackgammon -debug
run10: gobackgammon
	./gobackgammon -debug -auto -match 10
runauto: gobackgammon
	./gobackgammon -auto
# TODO(chandler37): Before using go 1.11 modules (see go.mod), this worked. Fix it.
doc:
	godoc -http=:6060
textdoc:
	go doc github.com/chandler37/gobackgammon/ai
	@echo " "
	go doc github.com/chandler37/gobackgammon/brd
	@echo " "
	go doc github.com/chandler37/gobackgammon/json
	@echo " "
	go doc github.com/chandler37/gobackgammon/svg
clean:
	rm -fr bin/ pkg/ coverprofile.out
	go clean -cache
	go clean ./...
