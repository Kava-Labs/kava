FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

VOLUME [ /kvd ]
WORKDIR /kvd
EXPOSE 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["rest-server", "--node", "http://192.168.10.3:26657", "--laddr", "tcp://0.0.0.0:1317"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh