FROM python:3.8.16-slim AS python

RUN useradd -ms /bin/bash newuser
WORKDIR /home/newuser

RUN mkdir -p ./scheduling/job/bestco/
WORKDIR ./scheduling
# copy only requirements.txt first for better caching
COPY ./internal/scheduling/job/bestco/requirements.txt ./job/bestco/
RUN pip3 install --no-cache-dir -r ./job/bestco/requirements.txt

COPY ./internal/scheduling .
RUN chmod -R 777 ./data

USER newuser
