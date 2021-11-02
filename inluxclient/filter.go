package influxclient

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
// 	client.BucketsAPI().Find(ctx context.Context, &influxapi.Filter{
//		OrgName: "example.com",
//		Limit: 10,
//		Offset: 40,
//	})
type Filter struct {
	// Filter by an organization name
	OrgName string
	// Filter by an organization ID
	OrgID string
	// Filter by a user ID
	UserID string
	// Maximum number of returned entities in a single request
	Limit uint
	// Starting offset for returning entities
	Offset uint
}
