module github.com/influxdata/influxdb-client-go/v2

go 1.13

require (
	github.com/deepmap/oapi-codegen v1.6.0
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1 // test dependency
	golang.org/x/net v0.0.0-20210119194325-5f4716e94777
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/deepmap/oapi-codegen v1.3.13 => github.com/bonitoo-io/oapi-codegen v1.3.8-0.20201014090437-baf71361141f
