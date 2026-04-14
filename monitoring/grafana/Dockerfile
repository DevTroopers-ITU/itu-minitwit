FROM grafana/grafana:12.1

ENV PROMETHEUS_HOST='host.docker.internal:9090'

COPY dashboards /etc/grafana/provisioning/dashboards
COPY datasources /etc/grafana/provisioning/datasources
COPY alerting /etc/grafana/provisioning/alerting
COPY entrypoint.sh /entrypoint.sh

USER root
RUN chmod +x /entrypoint.sh
USER grafana

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 3000
