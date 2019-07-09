ARG GO_IMAGE=gcr.io/cloud-builders/go:debian
FROM $GO_IMAGE as go

FROM gcr.io/cloud-builders/docker
ARG GOPATH=/workspace/go
ARG GOROOT=/usr/local/go

ENV GOPATH ${GOPATH}
ENV GOROOT ${GOROOT}

RUN mkdir -pv \
  /k8s-addon-builder \
  /builder \
  "${GOPATH}"

# Inject Golang.
COPY --from=go $GOROOT $GOROOT

ENV PATH="/k8s-addon-builder:/builder/google-cloud-sdk/bin:${GOROOT}/bin:${GOPATH}/bin:${PATH}"

# Compile "ply" binary statically.
WORKDIR /workspace/go/src/github.com/GoogleCloudPlatform/k8s-addon-builder
COPY / ./

RUN \
  # Install common build tools.
  apt-get update \
  && apt-get install -y --no-install-recommends \
    build-essential \
    git \
    make \
    wget \
    python-pip \
    python-yaml \
    unzip \
  # Install dep, which is needed by some addons.
  && go get -v -u github.com/golang/dep/cmd/dep \
  # Install ply (Golang).
  && make build-static \
  && cp ply builder-tools/* /k8s-addon-builder \
  # Install Python tools.
  && pip install -r /k8s-addon-builder/requirements.txt \
  && git config --system credential.helper gcloud.sh \
  && wget -q -O protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v3.7.1/protoc-3.7.1-linux-x86_64.zip \
  && unzip -p protoc.zip bin/protoc > /usr/local/bin/protoc \
  && chmod +x /usr/local/bin/protoc \
  && wget -qO- https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz | tar zxv -C /builder \
  && CLOUDSDK_PYTHON="python2.7" /builder/google-cloud-sdk/install.sh \
    --usage-reporting=false \
    --bash-completion=false \
    --disable-installation-options \
  # Clean up.
  && rm -rf \
    /workspace/go/src \
    /var/lib/apt/lists/* \
    ~/.config/gcloud

# Parent image's entrypoint is docker, reset it to minimize confusion.
ENTRYPOINT []
