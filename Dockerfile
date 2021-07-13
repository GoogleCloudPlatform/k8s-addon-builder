ARG GO_IMAGE
ARG GO_VERSION
FROM ${GO_IMAGE}:${GO_VERSION} as go

FROM launcher.gcr.io/google/ubuntu1804

RUN apt-get -y update && \
    apt-get -y install \
        apt-transport-https \
        ca-certificates \
        curl \
        make \
        software-properties-common && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    apt-key fingerprint 0EBFCD88 && \
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       bionic \
       stable" && \
    apt-get -y update && \
    apt-get -y install docker-ce docker-ce-cli

ARG GOPATH=/workspace/go
ARG GOROOT=/usr/local/go

ENV GOPATH ${GOPATH}
ENV GOROOT ${GOROOT}

# Need to install go packages in something other than /workspace/ because
# that gets overwritten by GCB, but the GOPATH needs to be /workspace/go
# because many addons expect artifacts of `go get/build/install/etc.` to
# persist across stages.
ARG INSTALL_GOPATH='/builder/go'
ENV INSTALL_GOPATH ${INSTALL_GOPATH}

RUN mkdir -pv \
  /k8s-addon-builder \
  /builder \
  "${GOPATH}" \
  "${INSTALL_GOPATH}"

# Inject Golang.
COPY --from=go $GOROOT $GOROOT

ENV PATH="/k8s-addon-builder:/builder/google-cloud-sdk/bin:${GOROOT}/bin:${GOPATH}/bin:${INSTALL_GOPATH}/bin:${PATH}"

RUN \
  # Install common build tools.
  apt-get update \
  && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    make \
    jq \
    wget \
    python-setuptools \
    python-pip \
    python-yaml \
    unzip

COPY ./builder-tools/requirements.txt /k8s-addon-builder/requirements.txt
RUN \
  # Install Python tools.
  pip install wheel \
  && pip install -r /k8s-addon-builder/requirements.txt \
  && git config --system credential.helper gcloud.sh \
  && wget -q -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip \
  && unzip -p protoc.zip bin/protoc > /usr/local/bin/protoc \
  && chmod +x /usr/local/bin/protoc \
  && unzip -o protoc.zip -d /usr/local include/* \
  && wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder \
  && CLOUDSDK_PYTHON="python2.7" /builder/google-cloud-sdk/install.sh \
    --usage-reporting=false \
    --bash-completion=false \
    --disable-installation-options

# Compile "ply" binary statically.
WORKDIR /workspace/go/src/github.com/GoogleCloudPlatform/k8s-addon-builder
COPY / ./
ARG KO_VERSION
ARG PLY_VERSION_GIT
ARG PLY_VERSION_DATE
ENV PLY_VERSION_GIT ${PLY_VERSION_GIT}
ENV PLY_VERSION_DATE ${PLY_VERSION_DATE}
RUN \
  # Install ply (Golang).
  make build-static \
  && cp ply /k8s-addon-builder \
  # preload ko, golang kubernetes builder
  && curl -L https://github.com/google/ko/releases/download/v${KO_VERSION}/ko_${KO_VERSION}_Linux_x86_64.tar.gz | tar xzf - ko \
  && chmod +x ./ko \
  && mv ko /bin

RUN \
  # Copy over builder-tools scripts to /k8s-addon-builder.
  cp builder-tools/* /k8s-addon-builder

RUN \
  # Clean up.
  rm -rf \
    /var/lib/apt/lists/* \
    ~/.config/gcloud

# Reset PWD to be something other than /workspace/... because it will get
# overridden to an empty volume when this image is invoked from Google Cloud
# Build (GCB).
WORKDIR /k8s-addon-builder

# Parent image's entrypoint is docker, reset it to minimize confusion.
ENTRYPOINT []
