## 1.5.0 [In progress]
1. [#165](https://github.com/influxdata/influxdb-client-go/pull/165) Allow overriding the http.Client for the http service.

## 1.4.0 [2020-07-17]
### Breaking changes
1. [#156](https://github.com/influxdata/influxdb-client-go/pull/156) Fixing Go naming and code style violations: 
- Introducing new *API interfaces with proper name of types, methods and arguments. 
- This also affects the `Client` interface and the `Options` type. 
- Affected types and methods have been deprecated and they will be removed in the next release. 

### Bug fixes 
1. [#152](https://github.com/influxdata/influxdb-client-go/pull/152) Allow connecting to server on a URL path
1. [#154](https://github.com/influxdata/influxdb-client-go/pull/154) Use idiomatic go style for write channels (internal)
1. [#155](https://github.com/influxdata/influxdb-client-go/pull/155) Fix panic in FindOrganizationByName in case of no permissions


## 1.3.0 [2020-06-19]
### Features
1. [#131](https://github.com/influxdata/influxdb-client-go/pull/131) Labels API
1. [#136](https://github.com/influxdata/influxdb-client-go/pull/136) Possibility to specify default tags
1. [#138](https://github.com/influxdata/influxdb-client-go/pull/138) Fix errors from InfluxDB 1.8 being empty

### Bug fixes 
1. [#132](https://github.com/influxdata/influxdb-client-go/pull/132) Handle unsupported write type as string instead of generating panic
1. [#134](https://github.com/influxdata/influxdb-client-go/pull/134) FluxQueryResult: support reordering of annotations

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
