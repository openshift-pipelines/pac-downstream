ARG GO_BUILDER=registry.access.redhat.com/ubi9/go-toolset:1.25
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:d91be7cea9f03a757d69ad7fcdfcd7849dba820110e7980d5e2a1f46ed06ea3b

FROM $GO_BUILDER AS builder

ARG TKN_PAC_VERSION=0.39
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
ARG VERSION=1.21

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel9-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:1.21::el9" \
    description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,cli" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel9" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    version="v1.21.1"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
