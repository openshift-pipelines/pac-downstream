ARG GO_BUILDER=registry.access.redhat.com/ubi9/go-toolset:latest@sha256:0a4666f7a4eb0644c97a73cba198eb268691b270d97831822689e7a2088f87be
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal@sha256:7b8e25a1b56ca4d00219198f3b5b51a3e1693a5c4f5369c5e190d7d6cb3f980e

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

FROM $RUNTIME
ARG VERSION=1.24

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel9-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:1.24::el9" \
    description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,cli" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel9" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    version="v1.24.1"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
