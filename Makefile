all: build done

dev: build run

build:
	@echo "Building..."
	@CGO_ENABLED=0 GOOS=${OS_NAME} GOARCH=amd64 go build -ldflags "-s -w -extldflags -static -X main.version=$(shell git rev-parse --short=8 HEAD)"
clean:
	@echo "Cleanup..."
	rm -f gopac
run:
	@echo "Running..."
	./gopac -debug
done:
	@echo "Done."
