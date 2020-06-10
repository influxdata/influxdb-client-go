## 1.3.0 [in progress]
### Features
1. [#131](https://github.com/influxdata/influxdb-client-go/pull/131) Labels API

## 1.2.0 [2020-05-15]
### Breaking Changes
 - [#107](https://github.com/influxdata/influxdb-client-go/pull/107) Renamed `InfluxDBClient` interface to `Client`, so the full name `influxdb2.Client` suits better to Go naming conventions
 - [#125](https://github.com/influxdata/influxdb-client-go/pull/125) `WriteApi`,`WriteApiBlocking`,`QueryApi` interfaces and related objects like `Point`, `FluxTableMetadata`, `FluxTableColumn`, `FluxRecord`, moved to the `api` ( and `api/write`, `api/query`) packages
 to provide consistent interface 
 
### Features
1. [#120](https://github.com/influxdata/influxdb-client-go/pull/120) Health check API   
1. [#122](https://github.com/influxdata/influxdb-client-go/pull/122) Delete API
1. [#124](https://github.com/influxdata/influxdb-client-go/pull/124) Buckets API

### Bug fixes 
1. [#108](https://github.com/influxdata/influxdb-client-go/issues/108) Fix default retry interval doc
1. [#110](https://github.com/influxdata/influxdb-client-go/issues/110) Allowing empty (nil) values in query result

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