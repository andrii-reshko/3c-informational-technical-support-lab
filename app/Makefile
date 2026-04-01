TFL_DIR = $(CURDIR)/libtflite

deps:
	docker build --target artifacts -f Tensorflow.dockerfile -o type=local,dest=./c .

collect:
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(TFL_DIR)/include" \
	CGO_LDFLAGS="-L$(TFL_DIR)/lib -Wl,-rpath,$(TFL_DIR)/lib" \
	go run -v ./cmd/collector/main.go

dev:
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(TFL_DIR)/include" \
	CGO_LDFLAGS="-L$(TFL_DIR)/lib -Wl,-rpath,$(TFL_DIR)/lib" \
	go run -v ./cmd/inference/main.go

run: dev

exp:
	CGO_ENABLED=0 \
	go run -v ./cmd/export_metrics/main.go

test:
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(TFL_DIR)/include" \
	CGO_LDFLAGS="-L$(TFL_DIR)/lib -Wl,-rpath,$(TFL_DIR)/lib" \
	go test -v ./...