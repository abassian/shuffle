BUILD_TAGS?=shl

# vendor uses Glide to install all the Go dependencies in vendor/
vendor:
	glide install

# install compiles and places the binary in GOPATH/bin
install:
	go install \
		--ldflags "-X github.com/abassian/shuffle/src/version.GitCommit=`git rev-parse HEAD` -X github.com/abassian/shuffle/src/version.GitBranch=`git symbolic-ref --short HEAD`" \
		./cmd/shl

test:
	glide novendor | xargs go test

.PHONY: vendor install test
