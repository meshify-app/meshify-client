FROM ubuntu:bionic

RUN mkdir -p /etc/meshify
RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf
RUN apt-get update && apt-get -y install curl gnupg
RUN curl -s -o /etc/apt/sources.list.d/meshify.list https://ppa.meshify.app/meshify.list
RUN curl https://ppa.meshify.app/meshify.gpg | gpg -o /usr/share/keyrings/meshify.gpg --dearmor --batch --yes
RUN apt-get update && apt-get -y install meshify-client wireguard-tools iproute2 inetutils-ping iptables
RUN apt-get install -y apt-utils debconf-utils dialog
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
RUN echo "resolvconf resolvconf/linkify-resolvconf boolean false" | debconf-set-selections
RUN apt-get update
RUN apt-get install -y resolvconf

CMD meshify-client


