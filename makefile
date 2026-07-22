.PHONY: run build test clean stop

run:
	go build -o out/musync && ./out/musync

build:
	go build -o out/musync

test:
	go test -v ./...

clean:
	rm -f out/musync

stop:
	killall musync
