FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/gitlab-operator \
    USER_UID=1001 \
    USER_NAME=gitlab-operator

# Copy config templates
COPY build/templates /templates

# install operator binary
COPY build/_output/bin/gitlab-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
