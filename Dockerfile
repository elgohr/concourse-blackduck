FROM openjdk:13
RUN adduser -m blackduck
ENV DETECT_JAR_PATH /home/blackduck
RUN /bin/bash -c "bash <(curl -s -L https://detect.synopsys.com/detect.sh) || true"
RUN chown -R blackduck ${DETECT_JAR_PATH} \
  && chmod +x ${DETECT_JAR_PATH}/hub-detect-java.sh
USER blackduck
WORKDIR ${DETECT_JAR_PATH}
