.PHONY: default
default: build

REPO_LOCAL=localhost/jouros
REPO_PUBLIC=registry.hub.docker.com/jrcjoro1/gorestapi
DOCKER_HUB=registry.hub.docker.com

.PHONY: build
build:
	podman build --format=docker --log-level=debug --tag $(REPO_PUBLIC):1.0 -f ./Dockerfile

.PHONY: push
push:
	podman push --log-level=debug $(REPO_PUBLIC):1.0


.PHONY: test
test:
	go test -cover ./... 

# use make git m="My comment"
.PHONY: git 
git:
	git add .	
	git commit -m "$m"
	git push origin main

dev:
	go run httpd/main.go
