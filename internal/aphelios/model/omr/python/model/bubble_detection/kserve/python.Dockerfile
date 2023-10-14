FROM python:3.8.13-slim AS py-developer

WORKDIR /

RUN apt-get update && apt-get install build-essential -y && apt-get install wget ffmpeg libsm6 libxext6 libpq-dev -y


# copy all source of python code, then build them.
WORKDIR /service
COPY mlserve.py .
COPY requirements.txt .

ENV RAY_DISABLE_MEMORY_MONITOR=1
RUN python -m pip install -r ./requirements.txt
ENTRYPOINT [ "/usr/local/bin/python", \
            "mlserve.py"]