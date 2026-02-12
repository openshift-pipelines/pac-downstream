ARG GO_BUILDER=registry.access.redhat.com/ubi9/go-toolset:9.7-1770596585@sha256:6983c6e7023236d1b5093c75c94402c7bdfd43c36ed9c34f2feda72a0ea7c8b1
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:759f5f42d9d6ce2a705e290b7fc549e2d2cd39312c4fa345f93c02e4abb8da95

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
ARG VERSION=pipelines-as-code-cli-next

COPY --from=builder /tmp/tkn-pac /usr/bin

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-cli-rhel9-container" \
      cpe="cpe:/a:redhat:openshift_pipelines:1.22::el9" \
      description="Red Hat OpenShift Pipelines pac-downstream cli" \
      io.k8s.description="Red Hat OpenShift Pipelines pac-downstream cli" \
      io.k8s.display-name="Red Hat OpenShift Pipelines pac-downstream cli" \
      io.openshift.tags="tekton,openshift,pac-downstream,cli" \
      maintainer="pipelines-extcomm@redhat.com" \
      name="openshift-pipelines/pipelines-pipelines-as-code-cli-rhel9" \
      summary="Red Hat OpenShift Pipelines pac-downstream cli" \
      version="v1.22.0"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532
