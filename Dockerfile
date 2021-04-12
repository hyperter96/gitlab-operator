# Build the manager binary
FROM golang:1.14 as builder
WORKDIR /workspace

# Copy in any existing Go cache, and download
# any missing dependencies.
ENV GOPATH=/go
COPY .go/pkg/mod/ /go/pkg/mod/
COPY go.mod go.sum ./

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY helm/ helm/
COPY controllers/ controllers/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

LABEL name=gitlab-operator \
      vendor='GitLab, Inc.' \
      description='Operator to deploy GitLab instances' \
      summary='GitLab is a DevOps lifecycle tool that provides Git repositories'

# Allow the chart directory to be overwritten with --build-arg
ARG CHART_DIR="/charts"

ENV USER_UID=1001 \
    HELM_CHARTS=${CHART_DIR}

# ADD GITLAB LICENSE
COPY LICENSE /licenses/GITLAB

# Copy config templates
COPY hack/assets /

# Add pre-packaged charts for the operator to deploy
COPY charts ${CHART_DIR}

WORKDIR /
COPY --from=builder /workspace/manager .
USER ${USER_UID}

ENTRYPOINT ["/manager"]
