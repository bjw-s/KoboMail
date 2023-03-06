default: build

build:
	GOOS=linux GOARCH=arm go build

clean:
	rm kobomail

release:
	goreleaser release --snapshot --clean
