ARG GO_BUILDER=registry.access.redhat.com/ubi9/go-toolset:9.7-1771417345@sha256:799cc027d5ad58cdc156b65286eb6389993ec14c496cf748c09834b7251e78dc
ARG RUNTIME=registry.access.redhat.com/ubi9/ubi-minimal:latest@sha256:c7d44146f826037f6873d99da479299b889473492d3c1ab8af86f08af04ec8a0

FROM $GO_BUILDER AS builder

WORKDIR /go/src/github.com/openshift-pipelines/pipelines-as-code
COPY upstream .
COPY .konflux/patches patches/
RUN set -e; for f in patches/*.patch; do echo ${f}; [[ -f ${f} ]] || continue; git apply ${f}; done
ENV GODEBUG="http2server=0"
COPY head HEAD
ENV GOEXPERIMENT="strictfipsruntime"
RUN go build -ldflags="-X 'knative.dev/pkg/changeset.rev=$(cat HEAD)'" -mod=vendor -tags disable_gcp,strictfipsruntime -v -o /tmp/pipelines-as-code-controller \
    ./cmd/pipelines-as-code-controller

FROM $RUNTIME
ARG VERSION=pipelines-as-code-controller-next

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-controller ${KO_APP}/pipelines-as-code-controller
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
      com.redhat.component="openshift-pipelines-pipelines-as-code-controller-rhel9-container" \
      cpe="cpe:/a:redhat:openshift_pipelines:1.22::el9" \
      description="Red Hat OpenShift Pipelines pac-downstream controller" \
      io.k8s.description="Red Hat OpenShift Pipelines pac-downstream controller" \
      io.k8s.display-name="Red Hat OpenShift Pipelines pac-downstream controller" \
      io.openshift.tags="tekton,openshift,pac-downstream,controller" \
      maintainer="pipelines-extcomm@redhat.com" \
      name="openshift-pipelines/pipelines-pipelines-as-code-controller-rhel9" \
      summary="Red Hat OpenShift Pipelines pac-downstream controller" \
      version="v1.22.0"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-controller"]
