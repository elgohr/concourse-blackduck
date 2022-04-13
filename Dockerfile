FROM golang:1.18.1 as buildResource
ENV GO111MODULE=on
WORKDIR /build/source
ADD resource ./
RUN go get -u github.com/maxbrunsfeld/counterfeiter/v6 \
 && go generate ./... \
 && go test -v ./... \
 && go build -o ../compiled/out out/out.go \
 && go build -o ../compiled/in in/in.go \
 && go build -o ../compiled/check check/check.go

FROM openjdk:19 as buildJava
RUN jlink --compress=2 \
   --no-man-pages \
   --module-path /opt/openjdk-13/jmods \
   --add-modules java.base,java.sql,java.desktop,java.naming \
--output /compressed

FROM debian:11.2 as runtime
ENV PATH=$PATH:/opt/jdk/bin
ENV DETECT_JAR_DOWNLOAD_DIR /opt/resource
COPY --from=buildJava /compressed /opt/jdk/
COPY --from=buildResource /build/compiled/* ${DETECT_JAR_DOWNLOAD_DIR}/
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
	ca-certificates \
	curl \
 && rm -rf /var/lib/apt/lists/*
RUN /bin/bash -c "bash <(curl -s -L https://detect.synopsys.com/detect.sh) || true"
RUN adduser --home /home/blackduck \
	--disabled-password \
	--gecos "Blackduck" \
	blackduck
RUN chown -R blackduck ${DETECT_JAR_DOWNLOAD_DIR}
RUN chmod +x ${DETECT_JAR_DOWNLOAD_DIR}/*
WORKDIR /
USER blackduck
