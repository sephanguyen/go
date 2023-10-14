FROM python:3.10.5-slim AS py-developer

WORKDIR /
RUN apt-get update && apt-get install build-essential -y && apt-get -y install libpq-dev gcc

# copy all source of python code, then build them.
WORKDIR /service

COPY ./python_requirements.txt ./python-env/
RUN python -m pip install --upgrade pip
RUN python -m pip install -r ./python-env/python_requirements.txt

ENTRYPOINT [ "/usr/local/bin/mlflow", "server", "--host=0.0.0.0"]

