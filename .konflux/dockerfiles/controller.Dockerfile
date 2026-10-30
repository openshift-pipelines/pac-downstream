ARG GO_BUILDER=registry.redhat.io/ubi9/go-toolset:latest@sha256:cab9089469c52029e4d8a4e9f717e87c51ab22eed448ebcada101b811f97dfe7
ARG UBI_RUNTIME=registry.redhat.io/ubi9/ubi-minimal-pqc:latest@sha256:72187ec58a94bd958ff48e41881d7c5d8bc0b49586cc1661da5ba4d92374c206

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

FROM $UBI_RUNTIME
ARG VERSION=1.24

ENV KO_APP=/ko-app \
    KO_DATA_PATH=/kodata

COPY --from=builder /tmp/pipelines-as-code-controller ${KO_APP}/pipelines-as-code-controller
COPY head ${KO_DATA_PATH}/HEAD

LABEL \
    com.redhat.component="openshift-pipelines-pipelines-as-code-controller-rhel9-container" \
    cpe="cpe:/a:redhat:openshift_pipelines:1.24::el9" \
    description="Red Hat OpenShift Pipelines pipelines-as-code controller" \
    io.k8s.description="Red Hat OpenShift Pipelines pipelines-as-code controller" \
    io.k8s.display-name="Red Hat OpenShift Pipelines pipelines-as-code controller" \
    io.openshift.tags="tekton,openshift,pipelines-as-code,controller" \
    maintainer="pipelines-extcomm@redhat.com" \
    name="openshift-pipelines/pipelines-pipelines-as-code-controller-rhel9" \
    summary="Red Hat OpenShift Pipelines pipelines-as-code controller" \
    version="v1.24.2"

RUN groupadd -r -g 65532 nonroot && \
    useradd --no-log-init -r -u 65532 -g nonroot nonroot
USER 65532

ENTRYPOINT ["/ko-app/pipelines-as-code-controller"]
