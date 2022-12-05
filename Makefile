.PHONY: test push build build_test run_test clean

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
