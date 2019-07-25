FROM ${BASE_IMAGE}

WORKDIR /opt

USER root

RUN mkdir -p /opt/app

COPY apiserver startup.sh /opt/

COPY jq /bin/

RUN chmod +x /opt/apiserver /opt/startup.sh /bin/jq

ENTRYPOINT ["./startup.sh"]
