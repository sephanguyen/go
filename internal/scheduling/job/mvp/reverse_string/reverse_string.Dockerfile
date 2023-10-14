FROM python:3.10.5-slim

RUN mkdir ./source
WORKDIR ./source
COPY . .

RUN pip3 install -r requirement.txt
CMD ["python3", "./reverse_string.py", "--input=Kubernetes"]
