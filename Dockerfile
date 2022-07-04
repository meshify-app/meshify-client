FROM ubuntu:jammy

RUN echo "nameserver 8.8.8.8" > /etc/resolv.conf
RUN apt-get update && apt-get -y install curl gnupg
RUN curl -s --compressed https://ppa.meshify.app/KEY.gpg | apt-key add -
RUN curl -s --compressed -o /etc/apt/sources.list.d/meshify.list https://ppa.meshify.app/meshify.list
RUN apt-get update && apt-get -y install meshify-client wireguard-tools iproute2 inetutils-ping iptables
RUN apt-get install -y apt-utils debconf-utils dialog
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections
RUN echo "resolvconf resolvconf/linkify-resolvconf boolean false" | debconf-set-selections
RUN apt-get update
RUN apt-get install -y resolvconf

CMD meshify-client


