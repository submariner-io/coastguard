FROM ubuntu:18.04

WORKDIR /var/coastguard

COPY coastguard-controller.sh /usr/local/bin

RUN chmod +x /usr/local/bin/coastguard-controller.sh

COPY coastguard-controller /usr/local/bin

# temporary sleep infinity so that we can do our debugging
ENTRYPOINT coastguard-controller.sh