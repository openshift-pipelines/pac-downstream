# Rebuild trigger: 1.15.4 release 2026-02-18
ARG GO_BUILDER=registry.access.redhat.com/ubi9/go-toolset:1.25
ARG RUNTIME=registry.access.redhat.com/ubi8/ubi-minimal:latest@sha256:6ed9271b3d3e7147728afaf8917a936c0db2f1badba59550288c2cc772afb58f

FROM $GO_BUILDER AS builder

ARG TKN_PAC_VERSION=0.27.2
WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
RUN go build -mod=vendor -tags disable_gcp -v  \
    -ldflags "-X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=${TKN_PAC_VERSION}" \
    -o /tmp/tkn-pac ./cmd/tkn-pac

FROM $RUNTIME
ARG VERSION=1.15

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel9-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:1.15::el9" \
    description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,cli" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel9" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    version="v1.15.5"

RUN microdnf install -y shadow-utils
RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot

USER 65532
# trigger rebuild 2026-02-14
