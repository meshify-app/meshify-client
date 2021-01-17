DIR := $(shell pwd)
SH := bash
PROD_IMAGE_NAME=meshify-client-build
USERID := $(shell id -u)
GROUPID := $(shell id -g)
export USERID GROUPID

VERSION=$(shell git describe --tags $(shell git rev-list --tags --max-count=1))

build:
	docker build .
	docker-compose up $(PROD_IMAGE_NAME)

build-deb:
	debchange -m -v $(VERSION) "Current git tag of meshify-client." && \
	./debian/rules clean binary && \
	mv ../*.deb $(DIR)/. && \
	./debian/rules clean && \
	git checkout -- ./debian/changelog
	chown $(USERID):$(GROUPID) *.deb

clean:
	-docker-compose down --rmi all
	-rm -rf *.deb
