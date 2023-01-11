FROM gcr.io/distroless/static:nonroot

COPY xeol-agent /usr/bin

USER nonroot:nobody

ARG BUILD_DATE
ARG BUILD_VERSION
ARG VCS_REF
ARG VCS_URL

LABEL org.opencontainers.image.created=$BUILD_DATE
LABEL org.opencontainers.image.title="xeol-agent"
LABEL org.opencontainers.image.description="The xeol-agent polls Kubernetes Cluster API(s) to tell xeol.io details about Kubernetes inventory (deployments, containers, pods, namespaces)"
LABEL org.opencontainers.image.source=$VCS_URL
LABEL org.opencontainers.image.revision=$VCS_REF
LABEL org.opencontainers.image.vendor="xeol Inc."
LABEL org.opencontainers.image.version=$BUILD_VERSION
LABEL org.opencontainers.image.licenses="Apache-2.0"

ENTRYPOINT ["xeol-agent"]
