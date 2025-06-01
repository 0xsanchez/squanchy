.PHONY: build release clean

build:
	@go build -o build/squanchy ./cmd/squanchy

release:
	@GOARCH=amd64 GOOS=windows CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_windows_amd64.exe ./cmd/squanchy/
	@GOARCH=arm64 GOOS=windows CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_windows_arm64.exe ./cmd/squanchy/
	@GOARCH=amd64 GOOS=darwin CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_macos_amd64 ./cmd/squanchy/
	@GOARCH=arm64 GOOS=darwin CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_macos_arm64 ./cmd/squanchy/
	@GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_linux_amd64 ./cmd/squanchy/
	@GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_linux_arm64 ./cmd/squanchy/
	@GOARCH=amd64 GOOS=freebsd CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_freebsd_amd64 ./cmd/squanchy/
	@GOARCH=arm64 GOOS=freebsd CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_freebsd_arm64 ./cmd/squanchy/
	@GOARCH=amd64 GOOS=openbsd CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_openbsd_amd64 ./cmd/squanchy/
	@GOARCH=arm64 GOOS=openbsd CGO_ENABLED=0 go build -trimpath -ldflags='-s -w -extldflags="-static"' -o build/squanchy_openbsd_arm64 ./cmd/squanchy/

clean:
	@if [ "build/*" ]; then rm build/*; fi