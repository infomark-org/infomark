.PHONY: build run

build:
	docker build . -t patwie/symphony:v1

run:
	docker run -v ${PWD}:/data \
		-it patwie/symphony  \
		/var/symphony/bin/symphony \
		-F /data/request.mod \
		-D /data/request.dat \
		-f /data/request.par