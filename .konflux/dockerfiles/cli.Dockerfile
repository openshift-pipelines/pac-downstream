ARG GO_BUILDER=registry.redhat.io/ubi9/go-toolset:latest@sha256:23c7cb17e9c53a65f9ab01f2edf1a4e07c9a9dc2258ed63d69e06925e7cea381
ARG UBI_RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:08d4c70f1baf6537a0728656cf7942d561e74a35916509d205b6519861266352

FROM $GO_BUILDER AS builder

ARG TKN_PAC_VERSION=nightly
WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
ENV GOEXPERIMENT="strictfipsruntime"
RUN go build -mod=vendor -tags disable_gcp,strictfipsruntime -v  \
    -ldflags "-X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=${TKN_PAC_VERSION}" \
    -o /tmp/tkn-pac ./cmd/tkn-pac

FROM $UBI_RUNTIME
ARG VERSION=1.23

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel9-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:1.23::el9" \
    description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,cli" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel9" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    version="v1.23.5"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
