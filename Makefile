PROJECT=$(shell basename $(shell pwd))
TAG=ghcr.io/johnjones4/${PROJECT}
VERSION=$(shell date +%s)

.PHONY: ui

all: info container ui

info:
	echo ${PROJECT} ${VERSION}

container:
	docker build --platform linux/x86_64 -t ${TAG} ./server
	docker push ${TAG}:latest
	docker image rm ${TAG}:latest

ui:
	tar zcvf ui.tar.gz ./ui/src
	git tag ${VERSION}
	git push origin ${VERSION}
	gh release create ${VERSION} ui.tar.gz --generate-notes

ci: container ui