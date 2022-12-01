test:
	@go test -v ./...


push:
	@git add -A && git commit -m "update" && git push origin master



.PHONY: test push
