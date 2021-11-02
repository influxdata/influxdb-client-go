# Generated types and API client

`oss.yml` is copied from InfluxDB and customized, until changes are merged. Must be periodically sync with latest changes
 and types and client must be re-generated


## Install oapi generator
`git clone git@github.com:vlastahajek/oapi-codegen.git`
`cd oapi-codegen`
`git checkout feat/template_helpers
`go install ./cmd/oapi-codegen/oapi-codegen.go`

## Download and sync latest swagger
`wget https://raw.githubusercontent.com/influxdata/openapi/master/contracts/oss.yml`
 
## Generate
`cd influxdb-client-go/influxclient/model`
 
Generate types
`oapi-codegen -generate types -o types.gen.go -package model oss.yml`

Generate client
`oapi-codegen -generate client -o client.gen.go -package model -templates .\templates oss.yml`

