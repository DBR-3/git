# Time-series 
Designed for forecasting time series based on period data


## Functionality

### Description




```mermaid
graph TD;
    Node_Exporter-->Prometheus;
    Prometheus-->Storge;
    Storge-->Web-server;
    Web-server-->Grafana;
```

##### Libraries