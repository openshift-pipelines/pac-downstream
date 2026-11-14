ARG GO_BUILDER=registry.access.redhat.com/ubi10/go-toolset:latest@sha256:05f3da136a4dbe077f30606e2e53a4f5cab185989c98309823ea06a9a1c52a50
ARG UBI_RUNTIME=registry.access.redhat.com/ubi10/ubi-minimal:latest@sha256:fb2f7f857cc60d88e3fad19da505083a682ca4569d7b55a61f64aaa43641f741

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
ARG VERSION=nightly

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel10-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:nightly::el10" \
    description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,cli" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel10" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code cli" \
    version="nightly"

# ubi10/ubi-minimal does not ship shadow-utils, which provides groupadd/useradd
RUN microdnf install -y shadow-utils && \
    groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
