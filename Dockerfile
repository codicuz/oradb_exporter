FROM oraclelinux:7-slim as base
ENV GODEBUG=cgocheck=0
ENV LD_LIBRARY_PATH=/usr/lib/oracle/12.2/client64/lib
COPY ["rootfs", "/"]
RUN yum -y install /distr/*.rpm; yum clean all; rm -rfv /distr

FROM base
COPY ["main", "/"]
CMD ["/main"]
