# Simple Telemetry Client (Proof of Concept)


This telemetry client collects and sends system information reports to an InfluxDB instance. 
It is designed to support multiple operating systems, with current compatibility for Linux and Windows.


## Configuration
To configure the client, provide a JSON file with the following structure:


```json
{
  "influx_token": "<Your InfluxDB Token>",
  "influx_url": "<InfluxDB URL>",
  "influx_org": "<Your Organization Name>",
  "influx_bucket": "<Bucket Name>",
  "influx_tags": {
    "host": "<Host Identifier Tag>",
    "client": "<Client Identifier Tag>"
  }
}
```

