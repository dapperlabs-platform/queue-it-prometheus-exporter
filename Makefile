export DOCKER_BUILDKIT=0
IMAGE_TAG?=$(shell git rev-parse --short HEAD)
IMAGE_NAME:=${REPO}/queue-it-prometheus-exporter

.PHONY: test
test:
	go vet -v ./...
	go test -failfast -race

.PHONY: build-and-push
build-and-push:
	go vet -v ./...
	docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .
	docker push ${IMAGE_NAME}:${IMAGE_TAG}
ifneq ("${RELEASE}", "")
	docker tag ${IMAGE_NAME}:${IMAGE_TAG} ${IMAGE_NAME}:latest
	docker push ${IMAGE_NAME}:latest
endif

