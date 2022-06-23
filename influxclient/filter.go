package influxclient

import "time"

// The Filter can be used to select the entities returned by the Find operations.
// Not all filter kinds are supported by all endpoints: see the individual API
// documentation for details.
//
// Note that Filter also provides support for paging (using Offset
// and Limit) - this is only useful with filters that might return more
// than one entity.
//
// For example, to a maximum of 10 buckets that have find all buckets that are
// associated with the organization "example.com" starting 40 entries into
// the results:
//
//	client.BucketsAPI().Find(ctx context.Context, &influxapi.Filter{
//		OrgName: "example.com",
//		Limit: 10,
//		Offset: 40,
//	})
type Filter struct {
	// Filter by resource ID
	ID string

	// Filter by resource name
	Name string

	// Filter by organization name
	OrgName string

	// Filter by organization ID
	OrgID string

	// Filter by user ID
	UserID string

	// Filter by user name
	UserName string

	// Filter by status
	Status string

	// Maximum number of returned entities in a single request
	Limit uint

	// Starting offset for returning entities
	Offset uint

	// After specifies that the search results should start after the item of given ID.
	After string

	// BeforeTime selects items to those scheduled before this time.
	BeforeTime time.Time

	// AfterTime selects items to those scheduled after this time.
	AfterTime time.Time
}
