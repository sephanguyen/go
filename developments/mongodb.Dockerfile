FROM docker.io/bitnami/mongodb:4.4.11-debian-10-r12 as mongodb-custom

USER root
ARG SOPS_VERSION=3.7.1
ADD https://github.com/mozilla/sops/releases/download/v${SOPS_VERSION}/sops-v${SOPS_VERSION}.linux /usr/local/bin/sops
RUN chmod 755 /usr/local/bin/sops

RUN touch mongodb.secrets.yaml
RUN chown 1001 mongodb.secrets.yaml

ENTRYPOINT [ "/opt/bitnami/scripts/mongodb/entrypoint.sh" ]
CMD [ "/opt/bitnami/scripts/mongodb/run.sh" ]