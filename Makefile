build:
	go build -gcflags="all=-N -l" -o bin/editor2 ./cmd && ./bin/editor2