FROM golang:1.12.3 as buildResource
WORKDIR /build
ADD resource ./source
RUN go test -v ./... \
  && go build -o compiled/out source/out/out.go \
  && go build -o compiled/in source/in/in.go \
  && go build -o compiled/check source/check/check.go \
  && chmod +x compiled/*

FROM openjdk:13 as buildJava
RUN jlink --compress=2 \
    --no-man-pages \
    --module-path /opt/openjdk-13/jmods \
    --add-modules java.base,java.sql \
--output /compressed

FROM debian:9.8-slim as runtime
ENV PATH=$PATH:/opt/jdk/bin
ENV DETECT_JAR_PATH /opt/resource
RUN adduser --home ${DETECT_JAR_PATH} --disabled-password --gecos "" blackduck
COPY --from=buildJava /compressed /opt/jdk/
COPY --from=buildResource /build/compiled/* ${DETECT_JAR_PATH}/
RUN apt-get update && apt-get install -y curl ca-certificates --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

WORKDIR ${DETECT_JAR_PATH}
RUN /bin/bash -c "bash <(curl -s -L https://detect.synopsys.com/detect.sh) || true"
RUN chown -R blackduck ${DETECT_JAR_PATH}
USER blackduck
