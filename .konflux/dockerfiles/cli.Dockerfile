ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.22
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:bafd57451de2daa71ed301b277d49bd120b474ed438367f087eac0b885a668dc

FROM $GO_BUILDER AS builder

ARG TKN_PAC_VERSION=0.32.0
WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
ENV GOEXPERIMENT=strictfipsruntime
RUN go build -mod=vendor -tags disable_gcp -tags strictfipsruntime -v  \
    -ldflags "-X github.com/openshift-pipelines/pipelines-as-code/pkg/params/version.Version=${TKN_PAC_VERSION}" \
    -o /tmp/tkn-pac ./cmd/tkn-pac

FROM $RUNTIME
ARG VERSION=pipelines-as-code-cli-1.18

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
      com.redhat.component="openshift-pipelines-cli-tkn-pac-container" \
      name="openshift-pipelines/pipelines-cli-tkn-pac-rhel9" \
      version=$VERSION \
      summary="Red Hat OpenShift pipelines tkn pac CLI" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="CLI client 'tkn-pac' for managing openshift pipelines" \
      io.k8s.display-name="Red Hat OpenShift Pipelines tkn pac CLI" \
      io.k8s.description="Red Hat OpenShift Pipelines tkn pac CLI" \
      io.openshift.tags="pipelines,tekton,openshift"

RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
