FROM centos:7

WORKDIR /opt

USER root

COPY controller /opt/controller

COPY startup.sh /opt/startup.sh

RUN chmod +x /opt/controller

RUN chmod +x /opt/startup.sh

ENTRYPOINT ["./startup.sh"]
