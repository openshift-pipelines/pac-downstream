ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.22
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:dee813b83663d420eb108983a1c94c614ff5d3fcb5159a7bd0324f0edbe7fca1

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp,strictfipsruntime -v -o /tmp/pipelines-as-code-controller \
    ./cmd/pipelines-as-code-controller

FROM $RUNTIME
ARG VERSION=pipelines-as-code-controller-main

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-controller ${KO_APP}/pipelines-as-code-controller
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-controller-container" \
      name="openshift-pipelines/pipelines-pipelines-as-code-controller-rhel8" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Pipelines as Code Controller" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Pipelines as Code Controller" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Pipelines as Code Controller" \
      io.k8s.description="Red Hat OpenShift Pipelines Pipelines as Code Controller" \
      io.openshift.tags="pipelines,tekton,openshift"

RUN microdnf install -y shadow-utils
RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-controller"]
