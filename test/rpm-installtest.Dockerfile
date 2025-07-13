ARG IMAGE=rockylinux/rockylinux:10

FROM ${IMAGE}
ARG PACKAGE=./yeet.rpm

COPY ${PACKAGE} ${PACKAGE}
RUN dnf -y install ${PACKAGE}
