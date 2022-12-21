VERSION="0.0.1"
BIN="bin/test"
BIN_WIN=".exe"
BIN_MACOS=""
ARCH="amd64"

.PHONY: test push build build_test run_test clean build_win build_darwin

test:
	@go test -v ./...


push:
	@git add -A && git commit -m "update" && git push origin master

build_test:
	@go test -c -v -race -timeout 1000s -run ^TestCmd$ github.com/realjf/utils

run_test:
	@./utils.test

# make tag t=<your_version>
tag:
	@echo '${t}'
	@git tag -a ${t} -m "${t}" && git push origin ${t}

dtag:
	@echo 'delete ${t}'
	@git push --delete origin ${t} && git tag -d ${t}


clean:
	@go clean -testcache

build:
	@env CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} go build -tags=linux -ldflags='-s -w -X main.Version=${VERSION}' -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o ${BIN}-linux-${ARCH}-${VERSION} ./example/test.go
	@echo 'done'

build_win:
	@env CGO_ENABLED=1 GOOS=windows GOARCH=${ARCH} go build -tags=windows -ldflags '-s -w -X main.Version=${VERSION}' -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o ${BIN}-windows-${ARCH}-${VERSION}${BIN_WIN} ./example/test.go
	@echo 'done'

build_darwin:
	@env CGO_ENABLED=1 GOOS=darwin GOARCH=${ARCH} go build -tags=darwin -ldflags '-s -w -X main.Version=${VERSION}' -gcflags="all=-trimpath=${PWD}" -asmflags="all=-trimpath=${PWD}" -o ${BIN}-darwin-${ARCH}-${VERSION}${BIN_MACOS} ./example/test.go
	@echo 'done'

