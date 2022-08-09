```mermaid
graph TD;
    Node_Exporter-->Prometheus;
    Prometheus-->Storge;
    Storge-->Web-server;
    Web-server-->Grafana;
``