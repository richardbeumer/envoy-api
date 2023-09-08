FROM mcr.microsoft.com/devcontainers/python:1-3.11-bullseye

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

COPY requirements.txt /tmp/requirements.txt
RUN pip3 install -r /tmp/requirements.txt

# Set the default shell to bash instead of sh
ENV SHELL /bin/bash