FROM debian:12.0

WORKDIR /backend

# install pg_dump
RUN apt-get update && apt-get -y install curl gnupg
RUN curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg
RUN echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt/ bookworm-pgdg main" | tee /etc/apt/sources.list.d/postgresql.list
RUN apt-get update && apt-get install -y postgresql-client-14


# A wait script to make schema_generator wait for postgres container
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/2.7.3/wait /wait
RUN chmod +x /wait

COPY ./build/gendbschema ./build/gendbschema
RUN chmod +x ./build/gendbschema

# Run wait, generator, then give write permission on host
CMD /wait && ./build/gendbschema --services=auth,eureka,bob,fatima,tom,zeus,invoicemgmt,entryexitmgmt,timesheet,mastermgmt,calendar,lessonmgmt,notificationmgmt \
    && chmod 777 /backend/mock/testing/testdata \
    && chmod 777 /backend/mock/testing/testdata/* \
    && chmod 666 /backend/mock/testing/testdata/*/* \
    && chmod 777 /backend/mock/testing/testdata/test/migrations
