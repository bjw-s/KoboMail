default: build

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build

clean:
	rm -f kobomail
	rm -rf dist

release: clean build
	mv kobomail KoboRoot/usr/local/kobomail/kobomail
	mkdir -p dist
	tar -cvzf dist/KoboRoot.tgz -C KoboRoot/ .
	rm KoboRoot/usr/local/kobomail/kobomail
