FROM python:3.10.5-slim AS py-developer

WORKDIR /

RUN apt-get update && apt-get install build-essential -y && apt-get install wget ffmpeg libsm6 libxext6 libpq-dev -y

# install sops golang
RUN mkdir -p sops

WORKDIR /sops
RUN wget https://github.com/mozilla/sops/releases/download/v3.7.3/sops_3.7.3_amd64.deb 
RUN dpkg -i sops_3.7.3_amd64.deb

# copy all source of python code, then build them.
WORKDIR /service

COPY ../../python_requirements.txt ./python-env/
RUN python -m pip install --upgrade pip
RUN python -m pip install -r ./python-env/python_requirements.txt

COPY ../../cmd/aphelios ./
COPY ../../internal/aphelios ./internal/aphelios
COPY ../../pkg/manabuf_py/aphelios/vision/v1 ./pkg/manabuf_py/aphelios/vision/v1

WORKDIR /service
ENTRYPOINT [ "/usr/local/bin/python", \
            "main.py"]
