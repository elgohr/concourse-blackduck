FROM golang:1.12.4 as buildResource
WORKDIR /build
ADD resource ./source
RUN go test -v ./... \
 && go build -o compiled/out source/out/out.go \
 && go build -o compiled/in source/in/in.go \
 && go build -o compiled/check source/check/check.go

FROM openjdk:13 as buildJava
RUN jlink --compress=2 \
   --no-man-pages \
   --module-path /opt/openjdk-13/jmods \
   --add-modules java.base,java.sql,java.desktop,java.naming \
--output /compressed

FROM debian:9.8 as runtime
ENV PATH=$PATH:/opt/jdk/bin
ENV DETECT_JAR_PATH /opt/resource
COPY --from=buildJava /compressed /opt/jdk/
COPY --from=buildResource /build/compiled/* ${DETECT_JAR_PATH}/
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
	ca-certificates \
	curl \
 && rm -rf /var/lib/apt/lists/*
RUN /bin/bash -c "cd ${DETECT_JAR_PATH}; bash <(curl -s -L https://detect.synopsys.com/detect.sh) || true"
RUN adduser --home /home/blackduck \
	--disabled-password \
	--gecos "Blackduck" \
	blackduck
RUN chown -R blackduck ${DETECT_JAR_PATH}
RUN chmod +x ${DETECT_JAR_PATH}/*
WORKDIR /home/blackduck
USER blackduck
