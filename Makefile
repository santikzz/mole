.PHONY: server client clean

server:
	cd server && go build -o ../bin/mole-server main.go

client:
	cd client && go build -o ../bin/mole main.go

all: server client

clean:
	rm -rf bin/

install-server: server
	sudo cp bin/mole-server /usr/local/bin/

install-client: client
	sudo cp bin/mole /usr/local/bin/

docker-server:
	docker build -f Dockerfile.server -t mole-server .