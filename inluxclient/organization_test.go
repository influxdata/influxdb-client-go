package influxclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/influxdata/influxdb-client-go/inluxclient/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationAPI_Find(t *testing.T) {
	var findTests = []struct {
		testName     string
		filter       *Filter
		response     string
		orgCount     int
		statusCode   int
		errorMessage string
	}{
		{
			testName: "Find 1 no filter",
			filter:   nil,
			response: `	{
		"links": {
		"next": "http://example.com",
			"prev": "http://example.com",
			"self": "http://example.com"
	},
		"orgs": [
	{
	"createdAt": "2019-08-24T14:15:22Z",
	"description": "string",
	"id": "123456",
	"links": {
	"buckets": "/api/v2/buckets?org=myorg",
	"dashboards": "/api/v2/dashboards?org=myorg",
	"labels": "/api/v2/orgs/1/labels",
	"members": "/api/v2/orgs/1/members",
	"owners": "/api/v2/orgs/1/owners",
	"secrets": "/api/v2/orgs/1/secrets",
	"self": "/api/v2/orgs/1",
	"tasks": "/api/v2/tasks?org=myorg"
	},
	"name": "my-org",
	"status": "active",
	"updatedAt": "2019-08-24T14:15:22Z"
	}
]
}`,
			orgCount:     1,
			statusCode:   200,
			errorMessage: "",
		},
		{
			testName: "Empty resultset",
			filter:   nil,
			response: `	{
		"links": {
		"next": "http://example.com",
			"prev": "http://example.com",
			"self": "http://example.com"
	},
		"orgs": []
}`,
			orgCount:     0,
			statusCode:   200,
			errorMessage: "",
		},
		{
			testName:     "Find nothing",
			filter:       nil,
			response:     `{"code": "not found", "message": "organization name \"myorg\" not found"}`,
			orgCount:     0,
			statusCode:   400,
			errorMessage: `cannot retrieve organizations: not found: organization name "myorg" not found`,
		},
	}
	for _, test := range findTests {
		t.Run(test.testName, func(t *testing.T) {
			ctx := context.Background()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				w.Header().Add("Content-Type", "application/json")
				w.WriteHeader(test.statusCode)
				if len(test.response) > 0 {
					w.Write([]byte(test.response))
				}
			}))
			defer server.Close()
			c, err := model.NewClientWithResponses(server.URL + "/api/v2/")
			require.NoError(t, err)
			orgsAPI := newOrganizationAPI(c)

			orgList, err := orgsAPI.Find(ctx, test.filter)
			if test.errorMessage != "" {
				require.Error(t, err)
				assert.Equal(t, test.errorMessage, err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, orgList)
				assert.Len(t, orgList, test.orgCount)
			}
		})
	}
}
