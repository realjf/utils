test:
	@go test -v ./...


push:
	@git add -A && git commit -m "update" && git push origin master

build_test:
	@go test -c -v -race -timeout 1000s -run ^TestCmd$ github.com/realjf/utils

run_test:
	@./utils.test


.PHONY: test push
