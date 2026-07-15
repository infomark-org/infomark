FROM ubuntu:18.04

RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  sudo \
  build-essential \
  unzip \
  openjdk-11-jdk \
  ant \
  python3 \
  ant-optional \
  && rm -rf /var/lib/apt/lists/*


COPY setup/ /app/setup/
COPY run.sh /app/run.sh
RUN chmod -R 705 /app

RUN groupadd -g 999 simple_ci_runner && \
    useradd -r -u 999 -g simple_ci_runner simple_ci_runner

RUN /app/setup/init.sh
USER simple_ci_runner

ENTRYPOINT ["/app/run.sh"]

