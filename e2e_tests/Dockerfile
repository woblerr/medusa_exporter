ARG CASSANDRA_VERSION="5.0.4"
ARG MEDUSA_BUILD_TYPE="release"
ARG MEDUSA_VERSION="0.25.1"
# For custom medusa builds
ARG MEDUSA_REPO_TAG="master"
ARG MEDUSA_REPO_URL="https://github.com/thelastpickle/cassandra-medusa"

FROM golang:1.24-bookworm AS builder
ARG REPO_BUILD_TAG
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 go build \
        -mod=vendor -trimpath \
        -ldflags "-s -w \
            -X github.com/prometheus/common/version.Version=${REPO_BUILD_TAG} \
            -X github.com/prometheus/common/version.BuildDate=$(date +%Y-%m-%dT%H:%M:%S%z) \
            -X github.com/prometheus/common/version.Branch=$(git rev-parse --abbrev-ref HEAD) \
            -X github.com/prometheus/common/version.Revision=$(git rev-parse --short HEAD) \
            -X github.com/prometheus/common/version.BuildUser=medusa_exporter" \
        -o medusa_exporter medusa_exporter.go

FROM cassandra:${CASSANDRA_VERSION} AS cassandra_base
ARG MEDUSA_BUILD_TYPE
ARG MEDUSA_VERSION
ARG MEDUSA_REPO_TAG
ARG MEDUSA_REPO_URL
USER root
RUN apt-get update -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        wget \
        ca-certificates \
        python3-pip \
        python3-setuptools \
        build-essential \
        openssh-server \
    && apt-get autoremove -y \
    && apt-get autopurge -y \
    && rm -rf /var/lib/apt/lists/*
RUN pip3 install --no-cache-dir --upgrade pip \
    && pip3 install --no-cache-dir \
        cassandra-driver \
        tenacity

FROM cassandra_base AS cassandra_medusa-release
ARG MEDUSA_VERSION
RUN pip3 install --no-cache-dir \
        cassandra-medusa==${MEDUSA_VERSION}

FROM cassandra_base AS cassandra_medusa-custom
ARG MEDUSA_REPO_TAG
ARG MEDUSA_REPO_URL
ENV POETRY_VERSION=2.2.1
RUN apt-get update -y \
    && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        gcc \
        g++ \
        make \
        libffi-dev \
        libssl-dev \
        python3-dev \
        python3-venv \
        git \
    && apt-get autoremove -y \
    && apt-get autopurge -y \
    && rm -rf /var/lib/apt/lists/*
RUN POETRY_NO_INTERACTION=1 \
    POETRY_VIRTUALENVS_IN_PROJECT=true \
    pip install --no-cache-dir poetry==${POETRY_VERSION}
RUN git clone --branch ${MEDUSA_REPO_TAG} ${MEDUSA_REPO_URL} /medusa \
    && cd /medusa \
    && poetry install \
    && poetry build \
    && cd /medusa/dist \
    && pip3 install --no-cache-dir cassandra_medusa-*.whl \
    && rm -rf /medusa

FROM cassandra_medusa-${MEDUSA_BUILD_TYPE} AS cassandra_medusa
ENV \
    # User from cassandra image
    EXPORTER_APP_USER="cassandra" \
    EXPORTER_PORT="19500" \
    EXPORTER_CONFIG=""
RUN mkdir -p \  
        /etc/medusa \
        /var/lib/cassandra/backups \
        /e2e_tests \
        /home/${EXPORTER_APP_USER} \
    && chown ${EXPORTER_APP_USER}:${EXPORTER_APP_USER} -R \
        /var/lib/cassandra \
        /e2e_tests \
        /home/${EXPORTER_APP_USER}
COPY e2e_tests/medusa.ini /etc/medusa/medusa.ini
COPY --chown=${EXPORTER_APP_USER}:${EXPORTER_APP_USER} --chmod=400 ./e2e_tests/server.* /e2e_tests/
COPY --chown=${EXPORTER_APP_USER}:${EXPORTER_APP_USER} --chmod=400 ./e2e_tests/user.* /e2e_tests/
COPY --chown=${EXPORTER_APP_USER}:${EXPORTER_APP_USER} --chmod=644 ./e2e_tests/web_config_*.yml /e2e_tests/
COPY --chown=${EXPORTER_APP_USER}:${EXPORTER_APP_USER} --chmod=755 ./e2e_tests/prepare_e2e.sh /e2e_tests/prepare_e2e.sh
COPY --from=builder --chmod=755 /build/medusa_exporter /usr/bin/medusa_exporter
USER ${EXPORTER_APP_USER}
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["/e2e_tests/prepare_e2e.sh", "${EXPORTER_CONFIG}"]
EXPOSE ${EXPORTER_PORT}