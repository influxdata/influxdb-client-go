# API Skeleton

Bellow is the skeleton for the new client V3 API with parts that are not implemented yet.
 
```go
// Package influxclient interacts with the InfluxDB network API.
//
// Various parts of the API are divided into sections by entity kind (for example organization,
// label, authorization), each implemented
// by a different Go type. The CRUD operations on each type follow the same pattern,
// defined by this generic interface:
//
//	type CRUD[Entity any] interface {
//		Find(ctx context.Context, filter *Filter) ([]Entity, error)
//		FindOne(ctx context.Context, filter *Filter) (Entity, error)
//		Create(ctx context.Context, entity Entity) (Entity, error)
//		Update(ctx context.Context, entity Entity) (Entity, error)
//		Delete(ctx context.Context, id string) error
//	}
//
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


package influxclient

// TODO define errors. what errors are defined by the API? eg. ErrNotFound etc.

//TODO Why batch size only 2,000? The recommended batch size is 5,000. 
const DefaultBatchSize = 2000

// Params holds the parameters for creating a new client.
// The only mandatory fields are ServerURL and AuthToken.
type Params struct {
	// ServerURL holds the URL of the InfluxDB server to connect to.
	// This must be non-empty.
	ServerURL string

	// AuthToken holds the authorization token for the API.
	// This can be obtained through the GUI web browser interface.
	AuthToken string

	// DefaultTags specifies a set of tags that will be added to each written
	// point. Tags specified on points override these.
	DefaultTags map[string]string

	// HTTPClient is used to make API requests.
	//
	// This can be used to specify a custom TLS configuration
	// (TLSClientConfig), a custom request timeout (Timeout),
	// or other customization as required.
	//
	// It HTTPClient is nil, http.DefaultClient will be used.
	HTTPClient *http.Client

	// BatchSize holds the default batch size
	// used by PointWriter. If it's zero, DefaultBatchSize will
	// be used. Note that this can be overridden with PointWriter.SetBatchSize.
	BatchSize int

	// FlushInterval holds the default flush interval used by PointWriter.
	// If it's zero, points must be flushed manually.
	// Note that this can be overridden with PointWriter.SetFlushInterval.
	FlushInterval time.Duration
}

// Client implements an InfluxDB client.
type Client struct{
	// unexported fields
}

// WritePoints writes all the given points to the server with the
// given organization id into the given bucket.
// The points are written synchronously. For a higher throughput
// API that buffers individual points and writes them asynchronously,
// use the PointWriter method.
func (c *Client) WritePoints(org, bucket string, points []*influxdata.Point) error

// PointWriter returns a PointWriter value that support fast asynchronous
// writing of points to Influx. All the points are written with the given organization
// id into the given bucket.
//
// The returned PointWriter must be closed after use to release resources
// and flush any buffered points.
func (c *Client) PointWriter(org, bucket string) *PointWriter

// DeletePoints deletes points from the given bucket with the given ID within the organization with
// the given organization ID that have timestamps between the two times provided.
//
// The predicate holds conditions for selecting the data for deletion.
// For example:
// 		tag1="value1" and (tag2="value2" and tag3!="value3")
// When predicate is empty, all points within the given time range will be deleted.
// See https://v2.docs.influxdata.com/v2.0/reference/syntax/delete-predicate/
// for more info about predicate syntax.
func (c *Client) DeletePoints(ctx context.Context, orgID, bucketID string, start, stop time.Time, predicate string) error

// Query sends the given flux query on the given organization ID.
// The result must be closed after use.
func (c *Client) Query(ctx context.Context, org, query string) (*QueryResult, error)

// Ready checks that the server is ready, and reports the duration the instance
// has been up if so. It does not validate authentication parameters.
// See https://docs.influxdata.com/influxdb/v2.0/api/#operation/GetReady.
func (c *Client) Ready() (time.Duration, error)

// Health returns an InfluxDB server health check result. Read the HealthCheck.Status field to get server status.
// Health doesn't validate authentication params.
func (c *Client) Health(ctx context.Context) (*HealthCheck, error)

// Authorization returns a value that can be used to interact with the
// authorization-related parts of the InfluxDB API.
func (c *Client) AuthorizationAPI() *AuthorizationAPI

// BucketAPI returns a value that can be used to interact with the
// bucket-related parts of the InfluxDB API.
func (c *Client) BucketAPI() *BucketAPI

// LabelsAPI returns Labels API client
func (c *Client) LabelsAPI() *LabelsAPI

// OrganizationAPI returns
func (c *Client) OrganizationAPI() *OrganizationAPI

// UsersAPI returns Users API client.
func (c *Client) UsersAPI() *UsersAPI

// TasksAPI returns Tasks API client.
func (c *Client) TasksAPI() *TasksAPI


// AuthorizationAPI holds methods related to authorization, as found under
// the /authorizations endpoint.
type AuthorizationAPI struct {
	// unexported fields.
}

// Find returns all Authorization records that satisfy the given filter.
func (a *AuthorizationAPI) Find(ctx context.Context, filter *Filter) ([]*Authorization, error)

// FindOne returns one authorization record that satisfies the given filter.
func (a *AuthorizationAPI) FindOne(ctx context.Context, filter *Filter) (*Authorization, error)

// Create creates an authorization record. The
func (a *AuthorizationAPI) Create(ctx context.Context, auth *Authorization) error

func (a *AuthorizationAPI) SetStatus(ctx context.Context, id authID, status AuthorizationStatus) error

func (a *DeleteAuthorization) Delete(ctx context.Context, id authID) error

type BucketAPI struct {
	// unexported fields
}

// Find returns all buckets matching the given filter.
func (a *BucketAPI) Find(ctx context.Context, filter *Filter) ([]*Bucket, error)

// FindOne returns one bucket that matches the given filter.
func (a *BucketAPI) FindOne(ctx context.Context, filter *Filter) (*Bucket, error)

// Create creates a bucket. Zero-valued fields in b will be filled in with defaults.
//
// Full information about the bucket is returned.
func (a *BucketAPI) Create(ctx context.Context, b *Bucket) (*Bucket, error)

// Update updates information about a bucket.
// The b.ID and b.OrgID fields must be specified.
func (a *BucketAPI) Update(ctx context.Context, b *Bucket) (*Bucket, error)

// Delete deletes the bucket with the given ID.
func (a *BucketAPI) Delete(ctx context.Context, bucketID string) error

// LabelAPI holds methods pertaining to label management.
type LabelAPI struct {
	// unexported fields
}

// Find returns all labels matching the given filter.
func (a *LabelAPI) Find(ctx context.Context, filter *Filter) ([]*Label, error)

// FindOne returns one label that matches the given filter.
func (a *LabelAPI) FindOne(ctx context.Context, filter *Filter) *Label, error)

// Create creates a new label with the given information.
// The label.Name field must be non-empty.
// The returned Label holds the ID of the new label.
func (a *LabelAPI) Create(label *Label) (*Label, error)

// Update updates the label's name and properties.
// The label.ID and label.OrgID fields must be set.
// If the name is empty, it won't be changed. If a property isn't mentioned, it won't be changed.
// A property can be removed by using an empty value for that property.
//
// Update returns the fully updated label.
// TODO would this be better if it took the id and orgID as separate args?
func (a *LabelAPI) Update(ctx context.Context, label *domain.Label) (*domain.Label, error)

// Delete deletes the label with the given ID.
func (a *LabelAPI) Delete(ctx context.Context, labelID string) error

// Find returns all organizations matching the given filter.
// Supported filters:
//	name, ID, userID
func (a *OrganizationAPI) Find(ctx context.Context, filter *Filter) ([]*Organization, error)

// FindOne returns one organization matching the given filter.
// Supported filters:
//	name
//	ID
//	userID
func (a *OrganizationAPI) FindOne(ctx context.Context, filter *Filter) (*Organization, error)

// Create creates a new organization. The returned Organization holds the new ID.
func (a  *OrganizationAPI) Create(ctx context.Context, org *Organization) (*Organization, error)

// Update updates information about the organization. The org.ID field must hold the ID
// of the organization to be changed.
func (a *OrganizationAPI) Update(ctx context.Context, org *Organization) (*Organization, error)

// Delete deletes the organization with the given ID.
func (a *OrganizationAPI) Delete(ctx context.Context, orgID string) error

// Members returns all members of the organization with the given ID.
func (a *OrganizationAPI) Members(ctx context.Context, orgID string) ([]ResourceMember, error)

// AddMember adds the user with the given ID to the organization with the given ID.
func (a *OrganizationAPI) AddMember(ctx context.Context, orgID, userID string) error

// AddMember removes the user with the given ID from the organization with the given ID.
func (a *OrganizationAPI) RemoveMember(ctx context.Context, orgID, userID string) error

// Owners returns all the owners of the organization with the given id.
func (a *OrganizationAPI) Owners(ctx context.Context, orgID string) ([]ResourceOwner, error)

// AddOwner adds an owner with the given userID to the organization with the given id.
func (a *OrganizationAPI) AddOwner(ctx context.Context, orgID, userID string) error

// Remove removes the user with the given userID from the organization with the given id.
func (a *OrganizationAPI) RemoveOwner(ctx context.Context, orgID, userID string) error

type UsersAPI struct {
	// unexported fields
}

// Find returns all users matching the given filter.
// Supported filters:
//	userID
//	id (same as userID)
//	username
//	name (same as username)
func (a *UsersAPI) Find(ctx context.Context, filter *Filter) ([]*User, error)

// Find returns one user that matches the given filter.
func (a *UsersAPI) FindOne(ctx context.Context, filter *Filter) (*User, error)

// Create creates a user. TODO specify which fields are required.
func (a *UsersAPI) Create(ctx context.Context, user *User) (*User, error)

// Update updates a user. The user.ID field must be specified.
// The complete user information is returned.
func (a *UsersAPI) Update(ctx context.Context, user *User) (*User, error)

// SetPassword sets the password for the user with the given ID.
func (a *UsersAPI) SetPassword(ctx context.Context, userID, password string) error

// SetMyPassword sets the password associated with the current user.
// The oldPassword parameter must match the previously set password
// for the user.
func (a *UsersAPI) SetMyPassword(ctx context.Context, oldPassword, newPassword string) error

// Delete deletes the user with the given ID.
func (a *UsersAPI) Delete(ctx context.Context, userID string) error

type TasksAPI struct {
	// unexported fields
}

// Find returns all tasks matching the given filter.
// Supported filters:
//  name
//  orgName
//	orgID
//	user
//  status
func (a *TasksAPI) Find(ctx context.Context, filter *Filter) ([]*Task, error)

// Find returns one task  that matches the given filter.
func (a *TasksAPI) FindOne(ctx context.Context, filter *Filter) (*Task, error)
// CreateTask creates a new task according the the task object.
// Set OrgId, Name, Description, Flux, Status and Every or Cron properties.
// Every and Cron are mutually exclusive. Every has higher priority.
func (a *TasksAPI) Create(ctx context.Context, task *Task) (*Task, error)

// Update updates a task. The task.ID field must be specified.
// The complete task information is returned.
func (a *TasksAPI) Update(ctx context.Context, task *Task) (*Task, error)

// Delete deletes the task with the given ID.
func (a *TasksAPI) Delete(ctx context.Context, taskID string) error

// FindRuns returns a task runs according the filter. 
// Supported filters:
//    BeforeTime
//    AfterTime
func (a *TasksAPI) FindRuns(ctx context.Context, taskID string, filter *filter) ([]*TaskRun, error)

// FindOneRun returns one task run that matches the given filter.
func (a *TasksAPI) FindOneRun(ctx context.Context, filter *Filter) (*TaskRun, error)

// FindRunLogs return all log events for a task run with given ID.
func (a *TasksAPI) FindRunLogs(ctx context.Context, runID string) ([]LogEvent, error)

// RunManually manually start a run of a task with given ID now, overriding the current schedule.
func (a *TasksAPI) RunManually(ctx context.Context, taskID string) (*TaskRun, error)

// CancelRun cancels a running task with given ID and given run ID.
func (a *TasksAPI) CancelRun(ctx context.Context, taskID, runID string) error

// RetryRun retry a run with given ID of a task with given ID.
func (a *TasksAPI) RetryRun(ctx context.Context, taskID, runID string) (*TaskRun, error)

// FindLogs retrieves all logs for a task with given ID.
func (a *TasksAPI) FindLogs(ctx context.Context, taskID string) ([]LogEvent, error)

// FindLabels retrieves labels of a task with given ID.
func (a *TasksAPI) FindLabels(ctx context.Context, taskID string) ([]Label, error)

// AddLabel adds a label with given ID to a task with given ID.
func (a *TasksAPI) AddLabel(ctx context.Context, taskID, labelID string) (*Label, error)

// RemoveLabel removes a label with given ID  from a task with given ID.
func (a *TasksAPI) RemoveLabel(ctx context.Context, taskID, labelID string) error

// Members returns all members of the task with the given ID.
func (a *TasksAPI) Members(ctx context.Context, taskID string) ([]ResourceMember, error)

// AddMember adds the user with the given ID to the task with the given ID.
func (a *TasksAPI) AddMember(ctx context.Context, taskID, userID string) error

// AddMember removes the user with the given ID from the task with the given ID.
func (a *TasksAPI) RemoveMember(ctx context.Context, taskID, userID string) error

// Owners returns all the owners of the task with the given id.
func (a *TasksAPI) Owners(ctx context.Context, taskID string) ([]ResourceOwner, error)

// AddOwner adds an owner with the given userID to the task with the given id.
func (a *TasksAPI) AddOwner(ctx context.Context, taskID, userID string) error

// Remove removes the user with the given userID from the task with the given id.
func (a *TasksAPI) RemoveOwner(ctx context.Context, taskID, userID string) error


// Filter specifies a filter that chooses only certain results from
// an API Find operation. The zero value of a filter (or a nil *Filter)
// selects everything with the default limit on result count.
// 
// Not all Find endpoints support all filter kinds (see the relevant endpoints
// for details).
type Filter struct {
    // ID filters items to one with given ID
    ID string

	// Limit specifies that the search results should be limited
	// to n results. If n is greater than the maximum allowed by the
	// API, multiple calls will be made to accumulate the desired
	// count.
	//
	// As a special case, if n is -1, there will be no limit
	// and all results will be accumulated into the same slice.
	//
	// When Limit isn't used, the limit will be the default limit
	// used by the API.
	Limit int
	
	// Offset specifies that the search results should start at
	// the given index.
	Offset int

    // After specifies that the search results should start after
    // the item of given ID.
	After string
	
	// UserName selects items associated with
	// the user with the given name.
	UserName string

	// UserID selects items associated with the user
	// with the given ID.
	UserID string

	// OrganizationName selects items associated with the
	// organization with the given name.
	OrganizationName string

	// OrganizationID selects items associated with the
	// organization with the given ID.
	OrganizationID string

	// Status selects items with given status
	Status string

	// BeforeTime selects items to those scheduled before this time.
	BeforeTime time.Time

	// AfterTime selects items to those scheduled after this time.
	AfterTime time.Time
}

type QueryResult struct {
	// unexported fields.
}

// NextTable advances to the next table in the result.
// Any remaining data in the current table is discarded.
//
// When there are no more tables, it returns false.
func (r *QueryResult) NextTable() bool

// NextRow advances to the next row in the current table.
// When there are no more rows in the current table, it
// returns false.
func (r *QueryResult) NextRow() bool

// Columns returns information on the columns in the current
// table. It returns nil if there is no current table (for example
// before NextTable has been called, or after NextTable returns false).
func (r *QueryResult) Columns() []TableColumn

// Err returns any error encountered. This should be called after NextTable
// returns false to check that all the results were correctly received.
func (r *QueryResult) Err() error

// Values returns the values in the current row.
// It returns nil if there is no current row.
// All rows in a table have the same number of values.
// The caller should not use the slice after NextRow
// has been called again, because it's re-used.
func (r *QueryResult) Values() []interface{}

// Decode decodes the current row into x, which should be
// a pointer to a struct. Columns in the row are decoded into
// appropriate fields in the struct, using the tag conventions
// described by encoding/json to determine how to map
// column names to struct fields.
func (r *QueryResult) Decode(x interface{}) error

// PointWriter implements a batching point writer.
type PointWriter struct {
	// unexported fields
}

// SetBatchSize sets the batch size associated with the writer, which
// determines the maximum number of points that will be written
// at any one point.
//
// This method must be called before the first call to WritePoint.
func (w *PointWriter) SetBatchSize(n int)

// SetFlushInterval sets the interval after which points will be
// flushed even when the buffer isn't full.
//
// This method must be called before the first call to WritePoint.
func (w *PointWriter) SetFlushInterval(time.Duration)

// WritePoint writes the given point to the API. As points are buffered,
// the error does not necessarily refer to the current point.
func (w *PointWriter) WritePoint(p *influxdata.Point) error

// Flush flushes all buffered points to the server.
func (w *PointWriter) Flush() error

// Close flushes all buffered points and closes the PointWriter.
// This must be called when the PointWriter is finished with.
//
// It's OK to call Close more than once - all calls will return the same error
// value.
func (w *PointWriter) Close() error
```
