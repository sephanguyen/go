# syntax=docker/dockerfile:1.3

FROM singlespa/import-map-deployer 
ENV PORT 5000

USER root
RUN chmod -R 775 /www


