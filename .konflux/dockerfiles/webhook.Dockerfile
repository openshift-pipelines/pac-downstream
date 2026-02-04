# Rebuild trigger: 1.15.4 release 2026-01-19
ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:ecd4751c45e076b4e1e8d37ac0b1b9c7271930c094d1bcc5e6a4d6954c6b2289

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp -v -o /tmp/pipelines-as-code-webhook \
    ./cmd/pipelines-as-code-webhook

FROM $RUNTIME
ARG VERSION=pipelines-as-code-webhook-1.15.3

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-webhook ${KO_APP}/pipelines-as-code-webhook
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-webhook-container" \
      name="openshift-pipelines/pipelines-pipelines-as-code-webhook-rhel8" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.k8s.description="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.openshift.tags="pipelines,tekton,openshift" \
      cpe="cpe:/a:redhat:openshift_pipelines:1.15::el8"

RUN microdnf install -y shadow-utils
RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-webhook"]
