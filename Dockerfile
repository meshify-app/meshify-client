FROM ubuntu:xenial

RUN apt-get update && apt-get -y install golang dh-golang devscripts git

ARG USERID
ARG GROUPID
WORKDIR /docker
CMD make -e build-deb
