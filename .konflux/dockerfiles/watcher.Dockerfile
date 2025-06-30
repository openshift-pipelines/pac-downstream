ARG GO_BUILDER=brew.registry.redhat.io/rh-osbs/openshift-golang-builder:v1.23
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:383329bf9c4f968e87e85d30ba3a5cb988a3bbde28b8e4932dcd3a025fd9c98c

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
ENV GOEXPERIMENT="strictfipsruntime"
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp -tags strictfipsruntime -v -o /tmp/pipelines-as-code-watcher \
    ./cmd/pipelines-as-code-watcher

FROM $RUNTIME
ARG VERSION=pipelines-as-code-watcher-1.19

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-watcher ${KO_APP}/pipelines-as-code-watcher
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-watcher-container" \
      name="openshift-pipelines/pipelines-as-code-watcher-rhel9" \
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
