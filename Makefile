all: build

build:
	go build -o himewiki ./cmd/himewiki

clean:
	rm -f himewiki

run:
	go run ./cmd/himewiki
