all: clean build

update_dep:
	go get $(DEP)
	go mod tidy

update_all_deps:
	go get -u
	go mod tidy

format:
	go fmt ./...

build:
	go build amievents2amqp.go

clean:
	rm -f amievents2amqp

build_images:
	docker build . --no-cache --target amievents2amqp-builder -t aleasoluciones/amievents2amqp-builder:${GIT_REV}
	docker build . --target amievents2amqp -t aleasoluciones/amievents2amqp:${GIT_REV}


.PHONY: update_dep update_all_deps format build clean build_images