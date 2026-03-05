ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:34880b64c07f28f64d95737f82f891516de9a3b43583f39970f7bf8e4cfa48b7

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
ENV GOEXPERIMENT=strictfipsruntime
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp -tags strictfipsruntime -v -o /tmp/pipelines-as-code-watcher \
    ./cmd/pipelines-as-code-watcher

FROM $RUNTIME
ARG VERSION=pipelines-as-code-watcher-1.18

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-watcher ${KO_APP}/pipelines-as-code-watcher
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-watcher-rhel9-container" \
      cpe="cpe:/a:redhat:openshift_pipelines:1.18::el9" \
      description="Red Hat OpenShift Pipelines pipelines-as-code watcher" \
      io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code watcher" \
      io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code watcher" \
      io.openshift.tags="tekton,openshift,pipelines-as-code,watcher" \
      maintainer="pipelines-extcomm@redhat.com" \
      name="openshift-pipelines/pipelines-pipelines-as-code-watcher-rhel9" \
      summary="Red Hat OpenShift Pipelines pipelines-as-code watcher" \
      version="v1.18.0"

RUN groupadd -r -g 65532 nonroot && useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-watcher"]
