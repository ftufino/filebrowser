FROM ghcr.io/ftufino/filebrowser:latest

USER root
COPY run.sh /run.sh
RUN chmod +x /run.sh
HEALTHCHECK NONE
ENTRYPOINT ["/run.sh"]
