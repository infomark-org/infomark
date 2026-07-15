FROM python:3.7

COPY setup/ /app/setup/
COPY run.sh /app/run.sh
RUN chmod -R 705 /app

RUN groupadd -g 999 simple_ci_runner && \
    useradd -r -u 999 -g simple_ci_runner simple_ci_runner

RUN /app/setup/init.sh

USER simple_ci_runner
ENTRYPOINT ["/app/run.sh"]