FROM redash/redash:10.1.0.b50633

USER root
RUN apt-get update && apt-get -y install netcat psmisc
RUN wget https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64 -O cloud_sql_proxy
RUN chmod +x cloud_sql_proxy && mv cloud_sql_proxy /cloud_sql_proxy

RUN wget https://storage.googleapis.com/alloydb-auth-proxy/v0.5.0/alloydb-auth-proxy.linux.amd64 -O alloydb-auth-proxy
RUN chmod +x alloydb-auth-proxy && mv alloydb-auth-proxy /alloydb-auth-proxy