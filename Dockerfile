FROM golang:1.12.3 as buildResource
WORKDIR /build
ADD resource ./source
RUN go test -v ./... \
  && go build -o compiled/out source/out/out.go \
  && go build -o compiled/in source/in/in.go \
  && go build -o compiled/check source/check/check.go \
  && chmod +x compiled/*

FROM openjdk:13
ENV DETECT_JAR_PATH /opt/resource
RUN mkdir -p ${DETECT_JAR_PATH} \
  && adduser -M -b ${DETECT_JAR_PATH} blackduck
RUN /bin/bash -c "bash <(curl -s -L https://detect.synopsys.com/detect.sh) || true"
COPY --from=buildResource /build/compiled/* ${DETECT_JAR_PATH}/
RUN chown -R blackduck ${DETECT_JAR_PATH} \
  && chmod +x ${DETECT_JAR_PATH}/*
USER blackduck
WORKDIR ${DETECT_JAR_PATH}
