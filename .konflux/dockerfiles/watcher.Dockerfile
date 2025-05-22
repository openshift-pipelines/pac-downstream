ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.redhat.io/ubi8/ubi:latest@sha256:244e9858f9d8a2792a3dceb850b4fa8fdbd67babebfde42587bfa919d5d1ecef

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp -v -o /tmp/pipelines-as-code-watcher \
    ./cmd/pipelines-as-code-watcher

FROM $RUNTIME
ARG VERSION=pipelines-as-code-watcher-1.16.4

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-watcher ${KO_APP}/pipelines-as-code-watcher
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-watcher-container" \
      name="openshift-pipelines/pipelines-as-code-watcher-rhel8" \
      version=$VERSION \
      summary="Red Hat OpenShift Pipelines Pipelines as Code Watcher" \
      maintainer="pipelines-extcomm@redhat.com" \
      description="Red Hat OpenShift Pipelines Pipelines as Code Watcher" \
      io.k8s.display-name="Red Hat OpenShift Pipelines Pipelines as Code Watcher" \
      io.k8s.description="Red Hat OpenShift Pipelines Pipelines as Code Watcher" \
      io.openshift.tags="pipelines,tekton,openshift"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-watcher"]
