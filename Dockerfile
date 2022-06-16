FROM ubuntu:jammy

WORKDIR /display

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update

RUN apt-get install curl python3 python3-pip python3-venv libbluetooth-dev libopenjp2-7 libtiff5 software-properties-common python3-pil python3-pil.imagetk -y

RUN pip install setuptools==58

COPY requirements.txt .

COPY lib lib

RUN pip install -r requirements.txt

COPY parsedata.py parsedata.py

COPY fonts fonts

COPY main.py main.py

COPY creds.json creds.json

RUN mkdir /temp

CMD ["python3", "main.py"]