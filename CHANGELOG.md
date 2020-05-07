## 1.2.0 [in progress]
### Features
1. [#120](https://github.com/influxdata/influxdb-client-go/pull/120) Health check API   
1. [#121](https://github.com/influxdata/influxdb-client-go/pull/121) Remove trailing slash from connection URL  

### Breaking Change
 - [#107](https://github.com/influxdata/influxdb-client-go/pull/100) Renamed `InfluxDBClient` interface to `Client`, so the full name `influxdb2.Client` suits better to Go naming conventions

### Bug fixes 
1. [#108](https://github.com/influxdata/influxdb-client-go/issues/108) Fix default retry interval doc
1. [#110](https://github.com/influxdata/influxdb-client-go/issues/110) Allowing empty (nil) values

### Documentation
 - [#112](https://github.com/influxdata/influxdb-client-go/pull/112) Clarify how to use client with InfluxDB 1.8+
 - [#115](https://github.com/influxdata/influxdb-client-go/pull/115) Doc and examples for reading write api errors 

## 1.1.0 [2020-04-24]
### Features
1. [#100](https://github.com/influxdata/influxdb-client-go/pull/100)  HTTP request timeout made configurable
1. [#99](https://github.com/influxdata/influxdb-client-go/pull/99)  Organizations API and Users API
1. [#96](https://github.com/influxdata/influxdb-client-go/pull/96)  Authorization API

### Docs
1. [#101](https://github.com/influxdata/influxdb-client-go/pull/101) Added examples to API docs

## 1.0.0 [2020-04-01]
### Core

- initial release of new client version

### Apis

- initial release of new client version