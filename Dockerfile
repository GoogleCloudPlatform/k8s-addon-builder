# Compile Go binary statically.
FROM k8s.gcr.io/addon-builder as builder
WORKDIR /workspace/go/src/github.com/GoogleCloudPlatform/addon-builder
COPY / ./
RUN dep ensure \
  && make build-static

# Extract static binary into scratch image. It could be any other x86 arch
# image.
FROM k8s.gcr.io/addon-builder
ENV PATH="/:${PATH}"
# Copy in entrypoint binaries into the toplevel root folder.
COPY --from=builder /workspace/go/src/github.com/GoogleCloudPlatform/addon-builder/cola /
