# Generated types and API client

`swagger.yml` is copied from InfluxDB and customized. Must be periodically sync with latest changes
 and types and client must be re-generated
## Install oapi generator
`git clone git@github.com:bonitoo-io/oapi-codegen.git`
`cd oapi-codegen`
`go build ./cmd/oapi-codegen/oapi-codegen.go` 
## Generate
`cd domain`
 
Generate types
`oapi-codegen -generate types -o types.gen.go -package domain swagger.yml`

Generate client
`oapi-codegen -generate client -o client.gen.go -package domain -templates .\templates swagger.yml`

