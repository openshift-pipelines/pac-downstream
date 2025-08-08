ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:8d905a93f1392d4a8f7fb906bd49bf540290674b28d82de3536bb4d0898bf9d7

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
ENV GOEXPERIMENT="strictfipsruntime"
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp,strictfipsruntime -v -o /tmp/pipelines-as-code-webhook \
    ./cmd/pipelines-as-code-webhook

FROM $RUNTIME
ARG VERSION=pipelines-as-code-webhook-next

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-webhook ${KO_APP}/pipelines-as-code-webhook
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-webhook-container" \
      name="openshift-pipelines/pipelines-pipelines-as-code-webhook-rhel9" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.k8s.description="Red Hat OpenShift Pipelines Pipelines as Code Webhook" \
      io.openshift.tags="pipelines,tekton,openshift"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-webhook"]
