FROM ghcr.io/ftufino/filebrowser:latest

USER root
COPY run.sh /run.sh
COPY healthcheck.sh /ha-healthcheck.sh
RUN chmod +x /run.sh /ha-healthcheck.sh
HEALTHCHECK --start-period=10s --interval=15s --timeout=3s CMD /ha-healthcheck.sh
ENTRYPOINT ["/run.sh"]
