FROM python:3.10.5-slim AS py-developer

WORKDIR /

RUN apt-get update && apt-get install build-essential -y && apt-get install wget ffmpeg libsm6 libxext6 libpq-dev -y

# install sops golang
RUN mkdir -p sops

WORKDIR /sops
RUN wget https://github.com/mozilla/sops/releases/download/v3.7.3/sops_3.7.3_amd64.deb
RUN dpkg -i sops_3.7.3_amd64.deb

#################################################
FROM py-developer
# copy all source of python code, then build them.
WORKDIR /service

# base python env
COPY ./python_requirements.txt ./python-env/
RUN python -m pip install --upgrade pip
RUN python -m pip install -r ./python-env/python_requirements.txt

# auto-scheduling package
COPY ./internal/scheduling/job/mvp/reverse_string/requirement.txt ./python-env/
RUN python -m pip install -r ./python-env/requirement.txt

COPY ./cmd/server/scheduling ./cmd/server/scheduling
COPY ./internal/scheduling ./internal/scheduling
COPY ./pkg/manabuf_py/scheduling/v1 ./pkg/manabuf_py/scheduling/v1

WORKDIR /service
EXPOSE 7280
ENTRYPOINT [ "/usr/local/bin/python", "./cmd/server/scheduling/http_server.py"]
