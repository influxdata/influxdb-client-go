# InfluxDB Client Go

[![CircleCI](https://circleci.com/gh/influxdata/influxdb-client-go/tree/v3.svg?style=svg)](https://circleci.com/gh/influxdata/influxdb-client-go/tree/v3)
[![codecov](https://codecov.io/gh/influxdata/influxdb-client-go/branch/v3/graph/badge.svg)](https://app.codecov.io/gh/influxdata/influxdb-client-go/branch/v3)

InfluxDB 2 Client Go v3 with new API. 

State: under development. 

[TODO](TODO.md) contains parts of API remaining for implementation.


| API | Endpoint | Status |
|:----------|:----------|:----------|
| `WritePoints()` | `/api/v2/write` |  | 
| `PointWriter` | `/api/v2/write` |  |
| `Query()` | `/api/v2/query` | In progress |
| `DeletePoints()` | `/api/v2/delete` |  |
| `Ready()` | `/ready` |  |
| `Health()` | `/health` |  |
| `AuthorizationAPI` | `/api/v2/authorizatons` |  |
| `BucketAPI` | `/api/v2/buckets` |  |
| `OrganizationAPI` | `/api/v2/orgs` |  |
| `UsersAPI` | `/api/v2/users` |  |
| `LabelsAPI` | `/api/v2/labels` |  |
| `TasksAPI` | `/api/v2/tasks` |  |

## License

The InfluxDB 2 Go Client is released under the [MIT License](https://opensource.org/licenses/MIT).
