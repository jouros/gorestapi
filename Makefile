.PHONY: default
default: build

REPO_LOCAL=localhost/jouros
REPO_PUBLIC=registry.hub.docker.com/jrcjoro1/gorestapi
DOCKER_HUB=registry.hub.docker.com

# use make build tag="TAG"
.PHONY: build
build:
	podman build --format=docker --log-level=debug --tag $(REPO_PUBLIC):$(tag) -f ./Dockerfile

# use make push tag="TAG", remember to do login first
.PHONY: push
push:
	podman push --log-level=debug $(REPO_PUBLIC):$(tag)

.PHONY: test
test:
	go test -cover ./... 

# use make git m="My comment"
.PHONY: git 
git:
	git add .	
	git commit -m "$m"
	git push origin main

.PHONY: dev
dev:
	go run main.go

# make podmanrun tag="TAG"
.PHONY: podmanrun
podmanrun:
	podman run -p 3000:3000 --log-level=debug registry.hub.docker.com/jrcjoro1/gorestapi:$(tag)
