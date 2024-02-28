# eero-metrics
>[!WARNING]
> This is a work in progress

This is a metrics exporter for eero devices connected to your network.

## Setup
Before you can start retreiving, you need to login and verify.

1. Run `metrics login <email|phone>` where `email|phone` is the email address or phone number you used to create your account. If successful you will be a sent code from Eero to the identifier you provided.  
This will print out your user token, that you should set as `EERO_USERTOKEN` environment variable for the next step.
2. Once you have the code, run `metrics validate <code>` where `<code>` is the code you received. If successful, you will receive a message and are good to run the next step.
3. Once you have validated your login, you can run `metrics serve` to start retrieving and publishing metrics. By default, metrics are available at `localhost:2112/metrics`

## Testing
You can use `docker-compose up` to run a local instance of Prometheus + Grafana + eero-metrics. 