export DOCKER_BUILDKIT=1
IMAGE_TAG?=$(shell git rev-parse --short HEAD)
IMAGE_NAME:=${REPO}/queue-it-prometheus-exporter

.PHONY: test
test:
	go vet -v ./...
	go test -failfast -race

.PHONY: build-local
build-local:
	go mod tidy
	CGO_ENABLED=0 go build -o ./queue-it-prometheus-exporter

.PHONY: build-and-push-image
build-and-push-image:
	go vet -v ./...
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
	docker push ${IMAGE_NAME}:${IMAGE_TAG}
ifeq ("${RELEASE}", "true")
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest
	docker push ${IMAGE_NAME}:latest
endif

