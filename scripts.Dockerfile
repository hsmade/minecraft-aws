FROM bitnami/aws-cli:2.7.35-debian-11-r0
USER root
RUN adduser --uid 1000 user
USER 1000
ADD scripts /scripts
