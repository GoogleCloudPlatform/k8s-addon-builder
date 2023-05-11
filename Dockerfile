ARG GO_IMAGE
ARG GO_VERSION
FROM ${GO_IMAGE}:${GO_VERSION} as go

FROM launcher.gcr.io/google/ubuntu2004

RUN apt-get -y update && \
    apt-get -y install \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        software-properties-common && \
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    apt-key fingerprint 0EBFCD88 && \
    add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       focal \
       stable" && \
    apt-get -y update && \
    apt-get -y install docker-ce=5:20.10.24~3-0~ubuntu-focal docker-ce-cli=5:20.10.24~3-0~ubuntu-focal

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
    python3-setuptools \
    python3-yaml \
    unzip

RUN \
  # Install gcloud sdk.
  echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" \
  | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list && \
  curl https://packages.cloud.google.com/apt/doc/apt-key.gpg \
  | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add - && \
  apt-get update -y && \
  apt-get install google-cloud-cli -y

RUN ln -s /usr/bin/python3 /usr/bin/python
RUN git config --system credential.helper gcloud.sh

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
