FROM centos:7

WORKDIR /opt

USER root

COPY apiserver /opt/apiserver

COPY startup.sh /opt/startup.sh

RUN chmod +x /opt/apiserver

RUN chmod +x /opt/startup.sh

ENTRYPOINT ["./startup.sh"]
