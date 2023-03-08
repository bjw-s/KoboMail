default: build

build:
	GOOS=linux GOARCH=arm go build

clean:
	rm -f kobomail
	rm -rf dist

release: clean build
	mkdir -p dist
	cp -r KoboRoot/* dist/
	mv kobomail dist/usr/local/kobomail/kobomail
	tar -cvzf KoboRoot.tgz -C dist/ .
	rm -rf dist/*
	mv KoboRoot.tgz dist/
