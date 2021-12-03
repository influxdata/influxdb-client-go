// Influx OSS API Service

import (
	"strings"
	"list"
)

info: {
	title:   *"Influx OSS API Service" | string
	version: *"2.0.0" | string
}
// Flux query to be analyzed.
#LanguageRequest: {
	// Flux query script to be analyzed
	query: string
	...
}

// Query influx using the Flux language
#Query: {
	extern?: #File

	// Query script to execute.
	query: string

	// The type of query. Must be "flux".
	type?: "flux"

	// Enumeration of key/value pairs that respresent parameters to be
	// injected into query (can only specify either this field or
	// extern and not both)
	params?: {
		...
	}
	dialect?: #Dialect

	// Specifies the time that should be reported as "now" in the
	// query. Default is the server's now time.
	now?: string
	...
}

// Query influx using the InfluxQL language
#InfluxQLQuery: {
	// InfluxQL query execute.
	query: string

	// The type of query. Must be "influxql".
	type?: "influxql"

	// Bucket is to be used instead of the database and retention
	// policy specified in the InfluxQL query.
	bucket?: string
	...
}

// Represents a complete package source tree.
#Package: {
	type?: #NodeType

	// Package import path
	path?: string

	// Package name
	package?: string

	// Package files
	files?: [...#File]
	...
}

// Represents a source from a single file
#File: {
	type?: #NodeType

	// The name of the file.
	name?:    string
	package?: #PackageClause

	// A list of package imports
	imports?: [...#ImportDeclaration]

	// List of Flux statements
	body?: [...#Statement]
	...
}

// Defines a package identifier
#PackageClause: {
	type?: #NodeType
	name?: #Identifier
	...
}

// Declares a package import
#ImportDeclaration: {
	type?: #NodeType
	as?:   #Identifier
	path?: #StringLiteral
	...
}

// The delete predicate request.
#DeletePredicateRequest: {
	// RFC3339Nano
	start: string

	// RFC3339Nano
	stop: string

	// InfluxQL-like delete statement
	predicate?: string
	...
}
#Node: #Expression | #Block

// Type of AST node
#NodeType: string

// A set of statements
#Block: {
	type?: #NodeType

	// Block body
	body?: [...#Statement]
	...
}
#Statement: #BadStatement | #VariableAssignment | #MemberAssignment | #ExpressionStatement | #ReturnStatement | #OptionStatement | #BuiltinStatement | #TestStatement

// A placeholder for statements for which no correct statement
// nodes can be created
#BadStatement: {
	type?: #NodeType

	// Raw source text
	text?: string
	...
}

// Represents the declaration of a variable
#VariableAssignment: {
	type?: #NodeType
	id?:   #Identifier
	init?: #Expression
	...
}

// Object property assignment
#MemberAssignment: {
	type?:   #NodeType
	member?: #MemberExpression
	init?:   #Expression
	...
}

// May consist of an expression that does not return a value and
// is executed solely for its side-effects
#ExpressionStatement: {
	type?:       #NodeType
	expression?: #Expression
	...
}

// Defines an expression to return
#ReturnStatement: {
	type?:     #NodeType
	argument?: #Expression
	...
}

// A single variable declaration
#OptionStatement: {
	type?:       #NodeType
	assignment?: #VariableAssignment | #MemberAssignment
	...
}

// Declares a builtin identifier and its type
#BuiltinStatement: {
	type?: #NodeType
	id?:   #Identifier
	...
}

// Declares a Flux test case
#TestStatement: {
	type?:       #NodeType
	assignment?: #VariableAssignment
	...
}
#Expression: #ArrayExpression | #DictExpression | #FunctionExpression | #BinaryExpression | #CallExpression | #ConditionalExpression | #LogicalExpression | #MemberExpression | #IndexExpression | #ObjectExpression | #ParenExpression | #PipeExpression | #UnaryExpression | #BooleanLiteral | #DateTimeLiteral | #DurationLiteral | #FloatLiteral | #IntegerLiteral | #PipeLiteral | #RegexpLiteral | #StringLiteral | #UnsignedIntegerLiteral | #Identifier

// Used to create and directly specify the elements of an array
// object
#ArrayExpression: {
	type?: #NodeType

	// Elements of the array
	elements?: [...#Expression]
	...
}

// Used to create and directly specify the elements of a
// dictionary
#DictExpression: {
	type?: #NodeType

	// Elements of the dictionary
	elements?: [...#DictItem]
	...
}

// A key/value pair in a dictionary
#DictItem: {
	type?: #NodeType
	key?:  #Expression
	val?:  #Expression
	...
}

// Function expression
#FunctionExpression: {
	type?: #NodeType

	// Function parameters
	params?: [...#Property]
	body?: #Node
	...
}

// uses binary operators to act on two operands in an expression
#BinaryExpression: {
	type?:     #NodeType
	operator?: string
	left?:     #Expression
	right?:    #Expression
	...
}

// Represents a function call
#CallExpression: {
	type?:   #NodeType
	callee?: #Expression

	// Function arguments
	arguments?: [...#Expression]
	...
}

// Selects one of two expressions, `Alternate` or `Consequent`,
// depending on a third boolean expression, `Test`
#ConditionalExpression: {
	type?:       #NodeType
	test?:       #Expression
	alternate?:  #Expression
	consequent?: #Expression
	...
}

// Represents the rule conditions that collectively evaluate to
// either true or false
#LogicalExpression: {
	type?:     #NodeType
	operator?: string
	left?:     #Expression
	right?:    #Expression
	...
}

// Represents accessing a property of an object
#MemberExpression: {
	type?:     #NodeType
	object?:   #Expression
	property?: #PropertyKey
	...
}

// Represents indexing into an array
#IndexExpression: {
	type?:  #NodeType
	array?: #Expression
	index?: #Expression
	...
}

// Allows the declaration of an anonymous object within a
// declaration
#ObjectExpression: {
	type?: #NodeType

	// Object properties
	properties?: [...#Property]
	...
}

// Represents an expression wrapped in parenthesis
#ParenExpression: {
	type?:       #NodeType
	expression?: #Expression
	...
}

// Call expression with pipe argument
#PipeExpression: {
	type?:     #NodeType
	argument?: #Expression
	call?:     #CallExpression
	...
}

// Uses operators to act on a single operand in an expression
#UnaryExpression: {
	type?:     #NodeType
	operator?: string
	argument?: #Expression
	...
}

// Represents boolean values
#BooleanLiteral: {
	type?:  #NodeType
	value?: bool
	...
}

// Represents an instant in time with nanosecond precision using
// the syntax of golang's RFC3339 Nanosecond variant
#DateTimeLiteral: {
	type?:  #NodeType
	value?: string
	...
}

// Represents the elapsed time between two instants as an int64
// nanosecond count with syntax of golang's time.Duration
#DurationLiteral: {
	type?: #NodeType

	// Duration values
	values?: [...#Duration]
	...
}

// Represents floating point numbers according to the double
// representations defined by the IEEE-754-1985
#FloatLiteral: {
	type?:  #NodeType
	value?: number
	...
}

// Represents integer numbers
#IntegerLiteral: {
	type?:  #NodeType
	value?: string
	...
}

// Represents a specialized literal value, indicating the left
// hand value of a pipe expression
#PipeLiteral: {
	type?: #NodeType
	...
}

// Expressions begin and end with `/` and are regular expressions
// with syntax accepted by RE2
#RegexpLiteral: {
	type?:  #NodeType
	value?: string
	...
}

// Expressions begin and end with double quote marks
#StringLiteral: {
	type?:  #NodeType
	value?: string
	...
}

// Represents integer numbers
#UnsignedIntegerLiteral: {
	type?:  #NodeType
	value?: string
	...
}

// A pair consisting of length of time and the unit of time
// measured. It is the atomic unit from which all duration
// literals are composed.
#Duration: {
	type?:      #NodeType
	magnitude?: int
	unit?:      string
	...
}

// The value associated with a key
#Property: {
	type?:  #NodeType
	key?:   #PropertyKey
	value?: #Expression
	...
}
#PropertyKey: #Identifier | #StringLiteral

// A valid Flux identifier
#Identifier: {
	type?: #NodeType
	name?: string
	...
}

// Dialect are options to change the default CSV output format;
// https://www.w3.org/TR/2015/REC-tabular-metadata-20151217/#dialect-descriptions
#Dialect: {
	// If true, the results will contain a header row
	header?: bool | *true

	// Separator between cells; the default is ,
	delimiter?: strings.MaxRunes(1) & strings.MinRunes(1) | *","

	// https://www.w3.org/TR/2015/REC-tabular-data-model-20151217/#columns
	annotations?: list.UniqueItems() & [..."group" | "datatype" | "default"]

	// Character prefixed to comment strings
	commentPrefix?: strings.MaxRunes(1) & strings.MinRunes(0) | *"#"

	// Format of timestamps
	dateTimeFormat?: "RFC3339" | "RFC3339Nano" | *"RFC3339"
	...
}
#AuthorizationUpdateRequest: null | bool | number | string | [...] | {
	// If inactive the token is inactive and requests using the token
	// will be rejected.
	status?: "active" | "inactive" | *"active"

	// A description of the token.
	description?: string
	...
}
#PostBucketRequest: null | bool | number | string | [...] | {
	orgID:          string
	name:           string
	description?:   string
	rp?:            string
	retentionRules: #RetentionRules
	schemaType?:    #SchemaType | *"implicit"
	...
}
#Bucket: null | bool | number | string | [...] | {
	links?: {
		// URL to retrieve labels for this bucket
		labels?: #Link

		// URL to retrieve members that can read this bucket
		members?: #Link

		// URL to retrieve parent organization for this bucket
		org?: #Link

		// URL to retrieve owners that can read and write to this bucket.
		owners?: #Link

		// URL for this bucket
		self?: #Link

		// URL to write line protocol for this bucket
		write?: #Link
		...
	}
	id?:            string
	type?:          "user" | "system" | *"user"
	name:           string
	description?:   string
	orgID?:         string
	rp?:            string
	schemaType?:    #SchemaType | *"implicit"
	createdAt?:     string
	updatedAt?:     string
	retentionRules: #RetentionRules
	labels?:        #Labels
	...
}
#Buckets: {
	links?: #Links
	buckets?: [...#Bucket]
	...
}

// Rules to expire or retain data. No rules means data never
// expires.
#RetentionRules: [...#RetentionRule]

// Updates to an existing bucket resource.
#PatchBucketRequest: {
	name?:           string
	description?:    string
	retentionRules?: #PatchRetentionRules
	...
}

// Updates to rules to expire or retain data. No rules means no
// updates.
#PatchRetentionRules: [...#PatchRetentionRule]

// Updates to a rule to expire or retain data.
#PatchRetentionRule: {
	type: "expire" | *"expire"

	// Duration in seconds for how long data will be kept in the
	// database. 0 means infinite.
	everySeconds?: int & >=0

	// Shard duration measured in seconds.
	shardGroupDurationSeconds?: int
	...
}
#RetentionRule: {
	type: "expire" | *"expire"

	// Duration in seconds for how long data will be kept in the
	// database. 0 means infinite.
	everySeconds: int & >=0

	// Shard duration measured in seconds.
	shardGroupDurationSeconds?: int
	...
}

// URI of resource.
#Link: string
#Links: {
	next?: #Link
	self:  #Link
	prev?: #Link
	...
}
#Logs: {
	events?: [...#LogEvent]
	...
}
#LogEvent: {
	// Time event occurred, RFC3339Nano.
	time?: string

	// A description of the event that occurred.
	message?: string

	// the ID of the task that logged
	runID?: string
	...
}
#Organization: null | bool | number | string | [...] | {
	links?: {
		self?:       #Link
		members?:    #Link
		owners?:     #Link
		labels?:     #Link
		secrets?:    #Link
		buckets?:    #Link
		tasks?:      #Link
		dashboards?: #Link
		...
	}
	id?:          string
	name:         string
	description?: string
	createdAt?:   string
	updatedAt?:   string

	// If inactive the organization is inactive.
	status?: "active" | "inactive" | *"active"
	...
}
#Organizations: {
	links?: #Links
	orgs?: [...#Organization]
	...
}
#PostOrganizationRequest: {
	name:         string
	description?: string
	...
}
#PatchOrganizationRequest: {
	// New name to set on the organization
	name?: string

	// New description to set on the organization
	description?: string
	...
}
#TemplateApply: {
	dryRun?:  bool
	orgID?:   string
	stackID?: string
	template?: {
		contentType?: string
		sources?: [...string]
		contents?: #Template
		...
	}
	templates?: [...{
		contentType?: string
		sources?: [...string]
		contents?: #Template
		...
	}]
	envRefs?: [string]: string | int | number | bool
	secrets?: [string]: string
	remotes?: [...{
		url:          string
		contentType?: string
		...
	}]
	actions?: [...{
		action?: "skipKind"
		properties?: {
			kind: #TemplateKind
			...
		}
		...
	} | {
		action?: "skipResource"
		properties?: {
			kind:                 #TemplateKind
			resourceTemplateName: string
			...
		}
		...
	}]
	...
}
#TemplateKind: "Bucket" | "Check" | "CheckDeadman" | "CheckThreshold" | "Dashboard" | "Label" | "NotificationEndpoint" | "NotificationEndpointHTTP" | "NotificationEndpointPagerDuty" | "NotificationEndpointSlack" | "NotificationRule" | "Task" | "Telegraf" | "Variable"
#TemplateExportByID: {
	stackID?: string
	orgIDs?: [...{
		orgID?: string
		resourceFilters?: {
			byLabel?: [...string]
			byResourceKind?: [...#TemplateKind]
			...
		}
		...
	}]
	resources?: [...{
		id:   string
		kind: #TemplateKind

		// if defined with id, name is used for resource exported by id.
		// if defined independently, resources strictly matching name are
		// exported
		name?: string
		...
	}]
	...
}
#TemplateExportByName: {
	stackID?: string
	orgIDs?: [...{
		orgID?: string
		resourceFilters?: {
			byLabel?: [...string]
			byResourceKind?: [...#TemplateKind]
			...
		}
		...
	}]
	resources?: [...{
		kind: #TemplateKind
		name: string
		...
	}]
	...
}
#Template: [...{
	apiVersion?: string
	kind?:       #TemplateKind
	meta?: {
		name?: string
		...
	}
	spec?: {
		...
	}
	...
}]
#TemplateEnvReferences: [...{
	// Field the environment reference corresponds too
	resourceField: string

	// Key identified as environment reference and is the key
	// identified in the template
	envRefKey: string

	// Value provided to fulfill reference
	value?:
		null | (string | int | number | bool)

	// Default value that will be provided for the reference when no
	// value is provided
	defaultValue?:
		null | (string | int | number | bool)
	...
}]
#TemplateSummary: {
	sources?: [...string]
	stackID?: string
	summary?: {
		buckets?: [...{
			id?:               string
			orgID?:            string
			kind?:             #TemplateKind
			templateMetaName?: string
			name?:             string
			description?:      string
			retentionPeriod?:  int
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		checks?: [...#CheckDiscriminator & {
			kind?:             #TemplateKind
			templateMetaName?: string
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		dashboards?: [...{
			id?:               string
			orgID?:            string
			kind?:             #TemplateKind
			templateMetaName?: string
			name?:             string
			description?:      string
			labelAssociations?: [...#TemplateSummaryLabel]
			charts?: [...#TemplateChart]
			envReferences?: #TemplateEnvReferences
			...
		}]
		labels?: [...#TemplateSummaryLabel]
		labelMappings?: [...{
			status?:                   string
			resourceTemplateMetaName?: string
			resourceName?:             string
			resourceID?:               string
			resourceType?:             string
			labelTemplateMetaName?:    string
			labelName?:                string
			labelID?:                  string
			...
		}]
		missingEnvRefs?: [...string]
		missingSecrets?: [...string]
		notificationEndpoints?: [...#NotificationEndpointDiscriminator & {
			kind?:             #TemplateKind
			templateMetaName?: string
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		notificationRules?: [...{
			kind?:                     #TemplateKind
			templateMetaName?:         string
			name?:                     string
			description?:              string
			endpointTemplateMetaName?: string
			endpointID?:               string
			endpointType?:             string
			every?:                    string
			offset?:                   string
			messageTemplate?:          string
			status?:                   string
			statusRules?: [...{
				currentLevel?:  string
				previousLevel?: string
				...
			}]
			tagRules?: [...{
				key?:      string
				value?:    string
				operator?: string
				...
			}]
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		tasks?: [...{
			kind?:             #TemplateKind
			templateMetaName?: string
			id?:               string
			name?:             string
			cron?:             string
			description?:      string
			every?:            string
			offset?:           string
			query?:            string
			status?:           string
			envReferences?:    #TemplateEnvReferences
			...
		}]
		telegrafConfigs?: [...#TelegrafRequest & {
			kind?:             #TemplateKind
			templateMetaName?: string
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		variables?: [...{
			kind?:             #TemplateKind
			templateMetaName?: string
			id?:               string
			orgID?:            string
			name?:             string
			description?:      string
			arguments?:        #VariableProperties
			labelAssociations?: [...#TemplateSummaryLabel]
			envReferences?: #TemplateEnvReferences
			...
		}]
		...
	}
	diff?: {
		buckets?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?: {
				name?:           string
				description?:    string
				retentionRules?: #RetentionRules
				...
			}
			old?: {
				name?:           string
				description?:    string
				retentionRules?: #RetentionRules
				...
			}
			...
		}]
		checks?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?:              #CheckDiscriminator
			old?:              #CheckDiscriminator
			...
		}]
		dashboards?: [...{
			stateStatus?:      string
			id?:               string
			kind?:             #TemplateKind
			templateMetaName?: string
			new?: {
				name?:        string
				description?: string
				charts?: [...#TemplateChart]
				...
			}
			old?: {
				name?:        string
				description?: string
				charts?: [...#TemplateChart]
				...
			}
			...
		}]
		labels?: [...{
			stateStatus?:      string
			kind?:             #TemplateKind
			id?:               string
			templateMetaName?: string
			new?: {
				name?:        string
				color?:       string
				description?: string
				...
			}
			old?: {
				name?:        string
				color?:       string
				description?: string
				...
			}
			...
		}]
		labelMappings?: [...{
			status?:                   string
			resourceType?:             string
			resourceID?:               string
			resourceTemplateMetaName?: string
			resourceName?:             string
			labelID?:                  string
			labelTemplateMetaName?:    string
			labelName?:                string
			...
		}]
		notificationEndpoints?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?:              #NotificationEndpointDiscriminator
			old?:              #NotificationEndpointDiscriminator
			...
		}]
		notificationRules?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?: {
				name?:            string
				description?:     string
				endpointName?:    string
				endpointID?:      string
				endpointType?:    string
				every?:           string
				offset?:          string
				messageTemplate?: string
				status?:          string
				statusRules?: [...{
					currentLevel?:  string
					previousLevel?: string
					...
				}]
				tagRules?: [...{
					key?:      string
					value?:    string
					operator?: string
					...
				}]
				...
			}
			old?: {
				name?:            string
				description?:     string
				endpointName?:    string
				endpointID?:      string
				endpointType?:    string
				every?:           string
				offset?:          string
				messageTemplate?: string
				status?:          string
				statusRules?: [...{
					currentLevel?:  string
					previousLevel?: string
					...
				}]
				tagRules?: [...{
					key?:      string
					value?:    string
					operator?: string
					...
				}]
				...
			}
			...
		}]
		tasks?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?: {
				name?:        string
				cron?:        string
				description?: string
				every?:       string
				offset?:      string
				query?:       string
				status?:      string
				...
			}
			old?: {
				name?:        string
				cron?:        string
				description?: string
				every?:       string
				offset?:      string
				query?:       string
				status?:      string
				...
			}
			...
		}]
		telegrafConfigs?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?:              #TelegrafRequest
			old?:              #TelegrafRequest
			...
		}]
		variables?: [...{
			kind?:             #TemplateKind
			stateStatus?:      string
			id?:               string
			templateMetaName?: string
			new?: {
				name?:        string
				description?: string
				args?:        #VariableProperties
				...
			}
			old?: {
				name?:        string
				description?: string
				args?:        #VariableProperties
				...
			}
			...
		}]
		...
	}
	errors?: [...{
		kind?:   #TemplateKind
		reason?: string
		fields?: [...string]
		indexes?: [...int]
		...
	}]
	...
}
#TemplateSummaryLabel: {
	id?:               string
	orgID?:            string
	kind?:             #TemplateKind
	templateMetaName?: string
	name?:             string
	properties?: {
		color?:       string
		description?: string
		...
	}
	envReferences?: #TemplateEnvReferences
	...
}
#TemplateChart: {
	xPos?:       int
	yPos?:       int
	height?:     int
	width?:      int
	properties?: #ViewProperties
	...
}
#Stack: {
	id?:        string
	orgID?:     string
	createdAt?: string
	events?: [...{
		eventType?:   string
		name?:        string
		description?: string
		sources?: [...string]
		resources?: [...{
			apiVersion?:       string
			resourceID?:       string
			kind?:             #TemplateKind
			templateMetaName?: string
			associations?: [...{
				kind?:     #TemplateKind
				metaName?: string
				...
			}]
			links?: {
				self?: string
				...
			}
			...
		}]
		urls?: [...string]
		updatedAt?: string
		...
	}]
	...
}
#Runs: {
	links?: #Links
	runs?: [...#Run]
	...
}
#Run: null | bool | number | string | [...] | {
	id?:     string
	taskID?: string
	status?: "scheduled" | "started" | "failed" | "success" | "canceled"

	// Time used for run's "now" option, RFC3339.
	scheduledFor?: string

	// An array of logs associated with the run.
	log?: [...#LogEvent]

	// Time run started executing, RFC3339Nano.
	startedAt?: string

	// Time run finished executing, RFC3339Nano.
	finishedAt?: string

	// Time run was manually requested, RFC3339Nano.
	requestedAt?: string
	links?: {
		self?:  string
		task?:  string
		retry?: string
		...
	}
	...
}
#RunManually: null | bool | number | string | [...] | {
	// Time used for run's "now" option, RFC3339. Default is the
	// server's now time.
	scheduledFor?:
		null | string
	...
}
#Tasks: {
	links?: #Links
	tasks?: [...#Task]
	...
}
#Task: {
	id: string

	// The type of task, this can be used for filtering tasks on list
	// actions.
	type?: string

	// The ID of the organization that owns this Task.
	orgID: string

	// The name of the organization that owns this Task.
	org?: string

	// The name of the task.
	name: string

	// The ID of the user who owns this Task.
	ownerID?: string

	// An optional description of the task.
	description?: string
	status?:      #TaskStatusType
	labels?:      #Labels

	// The ID of the authorization used when this task communicates
	// with the query engine.
	authorizationID?: string

	// The Flux script to run for this task.
	flux: string

	// A simple task repetition schedule; parsed from Flux.
	every?: string

	// A task repetition schedule in the form '* * * * * *'; parsed
	// from Flux.
	cron?: string

	// Duration to delay after the schedule, before executing the
	// task; parsed from flux, if set to zero it will remove this
	// option and use 0 as the default.
	offset?: string

	// Timestamp of latest scheduled, completed run, RFC3339.
	latestCompleted?: string
	lastRunStatus?:   "failed" | "success" | "canceled"
	lastRunError?:    string
	createdAt?:       string
	updatedAt?:       string
	links?: {
		self?:    #Link
		owners?:  #Link
		members?: #Link
		runs?:    #Link
		logs?:    #Link
		labels?:  #Link
		...
	}
	...
}
#TaskStatusType: "active" | "inactive"
#UserResponse:   null | bool | number | string | [...] | {
	id?:      string
	oauthID?: string
	name:     string

	// If inactive the user is inactive.
	status?: "active" | "inactive" | *"active"
	links?: {
		self?: string
		...
	}
	...
}
#Flags: {
	...
}
#ResourceMember: #UserResponse & {
	role?: "member" | *"member"
	...
}
#ResourceMembers: {
	links?: {
		self?: string
		...
	}
	users?: [...#ResourceMember]
	...
}
#ResourceOwner: #UserResponse & {
	role?: "owner" | *"owner"
	...
}
#ResourceOwners: {
	links?: {
		self?: string
		...
	}
	users?: [...#ResourceOwner]
	...
}
#FluxSuggestions: {
	funcs?: [...#FluxSuggestion]
	...
}
#FluxSuggestion: {
	name?: string
	params?: [string]: string
	...
}
#Routes: null | bool | number | string | [...] | {
	authorizations?: string
	buckets?:        string
	dashboards?:     string
	external?: {
		statusFeed?: string
		...
	}
	variables?: string
	me?:        string
	flags?:     string
	orgs?:      string
	query?: {
		self?:        string
		ast?:         string
		analyze?:     string
		suggestions?: string
		...
	}
	setup?:   string
	signin?:  string
	signout?: string
	sources?: string
	system?: {
		metrics?: string
		debug?:   string
		health?:  string
		...
	}
	tasks?:     string
	telegrafs?: string
	users?:     string
	write?:     string
	...
}
#Error: null | bool | number | string | [...] | {
	// code is the machine-readable error code.
	code: "internal error" | "not found" | "conflict" | "invalid" | "unprocessable entity" | "empty value" | "unavailable" | "forbidden" | "too many requests" | "unauthorized" | "method not allowed" | "request too large" | "unsupported media type"

	// message is a human-readable message.
	message: string

	// op describes the logical code operation during error. Useful
	// for debugging.
	op?: string

	// err is a stack of errors that occurred during processing of the
	// request. Useful for debugging.
	err?: string
	...
}
#LineProtocolError: null | bool | number | string | [...] | {
	// Code is the machine-readable error code.
	code: "internal error" | "not found" | "conflict" | "invalid" | "empty value" | "unavailable"

	// Message is a human-readable message.
	message: string

	// Op describes the logical code operation during error. Useful
	// for debugging.
	op: string

	// Err is a stack of errors that occurred during processing of the
	// request. Useful for debugging.
	err: string

	// First line within sent body containing malformed data
	line?: int
	...
}
#LineProtocolLengthError: null | bool | number | string | [...] | {
	// Code is the machine-readable error code.
	code: "invalid"

	// Message is a human-readable message.
	message: string

	// Max length in bytes for a body of line-protocol.
	maxLength: int
	...
}
#Field: {
	// value is the value of the field. Meaning of the value is
	// implied by the `type` key
	value?: string

	// `type` describes the field type. `func` is a function. `field`
	// is a field reference.
	type?: "func" | "field" | "integer" | "number" | "regex" | "wildcard"

	// Alias overrides the field name in the returned response.
	// Applies only if type is `func`
	alias?: string

	// Args are the arguments to the function
	args?: [...#Field]
	...
}
#BuilderConfig: {
	buckets?: [...string]
	tags?: [...#BuilderTagsType]
	functions?: [...#BuilderFunctionsType]
	aggregateWindow?: {
		period?:     string
		fillValues?: bool
		...
	}
	...
}
#BuilderTagsType: {
	key?: string
	values?: [...string]
	aggregateFunctionType?: #BuilderAggregateFunctionType
	...
}
#BuilderAggregateFunctionType: "filter" | "group"
#BuilderFunctionsType: {
	name?: string
	...
}
#DashboardQuery: {
	// The text of the Flux query.
	text?:          string
	editMode?:      #QueryEditMode
	name?:          string
	builderConfig?: #BuilderConfig
	...
}
#QueryEditMode: "builder" | "advanced"

// The description of a particular axis for a visualization.
#Axis: {
	// The extents of an axis in the form [lower, upper]. Clients
	// determine whether bounds are to be inclusive or exclusive of
	// their limits
	bounds?: list.MaxItems(2) & [...] & [...string]

	// Label is a description of this Axis
	label?: string

	// Prefix represents a label prefix for formatting axis values.
	prefix?: string

	// Suffix represents a label suffix for formatting axis values.
	suffix?: string

	// Base represents the radix for formatting axis values.
	base?:  "" | "2" | "10"
	scale?: #AxisScale
	...
}

// Scale is the axis formatting scale. Supported: "log", "linear"
#AxisScale: "log" | "linear"

// Defines an encoding of data value into color space.
#DashboardColor: {
	// The unique ID of the view color.
	id: string

	// Type is how the color is used.
	type: "min" | "max" | "threshold" | "scale" | "text" | "background"

	// The hex number of the color
	hex: strings.MaxRunes(7) & strings.MinRunes(7)

	// The user-facing name of the hex color.
	name: string

	// The data value mapped to this color.
	value: number
	...
}

// Describes a field that can be renamed and made visible or
// invisible.
#RenamableField: {
	// The calculated name of a field.
	internalName?: string

	// The name that a field is renamed to by the user.
	displayName?: string

	// Indicates whether this field should be visible on the table.
	visible?: bool
	...
}
#XYViewProperties: {
	timeFormat?: string
	type:        "xy"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	axes:              #Axes
	staticLegend?:     #StaticLegend
	xColumn?:          string
	generateXAxisTicks?: [...string]
	xTotalTicks?: int
	xTickStart?:  number
	xTickStep?:   number
	yColumn?:     string
	generateYAxisTicks?: [...string]
	yTotalTicks?:                int
	yTickStart?:                 number
	yTickStep?:                  number
	shadeBelow?:                 bool
	hoverDimension?:             "auto" | "x" | "y" | "xy"
	position:                    "overlaid" | "stacked"
	geom:                        #XYGeom
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#XYGeom: "line" | "step" | "stacked" | "bar" | "monotoneX"
#BandViewProperties: {
	timeFormat?: string
	type:        "band"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	axes:              #Axes
	staticLegend?:     #StaticLegend
	xColumn?:          string
	generateXAxisTicks?: [...string]
	xTotalTicks?: int
	xTickStart?:  number
	xTickStep?:   number
	yColumn?:     string
	generateYAxisTicks?: [...string]
	yTotalTicks?:                int
	yTickStart?:                 number
	yTickStep?:                  number
	upperColumn?:                string
	mainColumn?:                 string
	lowerColumn?:                string
	hoverDimension?:             "auto" | "x" | "y" | "xy"
	geom:                        #XYGeom
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#LinePlusSingleStatProperties: {
	timeFormat?: string
	type:        "line-plus-single-stat"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	axes:              #Axes
	staticLegend?:     #StaticLegend
	xColumn?:          string
	generateXAxisTicks?: [...string]
	xTotalTicks?: int
	xTickStart?:  number
	xTickStep?:   number
	yColumn?:     string
	generateYAxisTicks?: [...string]
	yTotalTicks?:                int
	yTickStart?:                 number
	yTickStep?:                  number
	shadeBelow?:                 bool
	hoverDimension?:             "auto" | "x" | "y" | "xy"
	position:                    "overlaid" | "stacked"
	prefix:                      string
	suffix:                      string
	decimalPlaces:               #DecimalPlaces
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#MosaicViewProperties: {
	timeFormat?: string
	type:        "mosaic"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...string]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	xColumn:           string
	generateXAxisTicks?: [...string]
	xTotalTicks?:           int
	xTickStart?:            number
	xTickStep?:             number
	yLabelColumnSeparator?: string
	yLabelColumns?: [...string]
	ySeriesColumns: [...string]
	fillColumns: [...string]
	xDomain:                     list.MaxItems(2) & [...number]
	yDomain:                     list.MaxItems(2) & [...number]
	xAxisLabel:                  string
	yAxisLabel:                  string
	xPrefix:                     string
	xSuffix:                     string
	yPrefix:                     string
	ySuffix:                     string
	hoverDimension?:             "auto" | "x" | "y" | "xy"
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#ScatterViewProperties: {
	timeFormat?: string
	type:        "scatter"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...string]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	xColumn:           string
	generateXAxisTicks?: [...string]
	xTotalTicks?: int
	xTickStart?:  number
	xTickStep?:   number
	yColumn:      string
	generateYAxisTicks?: [...string]
	yTotalTicks?: int
	yTickStart?:  number
	yTickStep?:   number
	fillColumns: [...string]
	symbolColumns: [...string]
	xDomain:                     list.MaxItems(2) & [...number]
	yDomain:                     list.MaxItems(2) & [...number]
	xAxisLabel:                  string
	yAxisLabel:                  string
	xPrefix:                     string
	xSuffix:                     string
	yPrefix:                     string
	ySuffix:                     string
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#HeatmapViewProperties: {
	timeFormat?: string
	type:        "heatmap"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...string]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	xColumn:           string
	generateXAxisTicks?: [...string]
	xTotalTicks?: int
	xTickStart?:  number
	xTickStep?:   number
	yColumn:      string
	generateYAxisTicks?: [...string]
	yTotalTicks?:                int
	yTickStart?:                 number
	yTickStep?:                  number
	xDomain:                     list.MaxItems(2) & [...number]
	yDomain:                     list.MaxItems(2) & [...number]
	xAxisLabel:                  string
	yAxisLabel:                  string
	xPrefix:                     string
	xSuffix:                     string
	yPrefix:                     string
	ySuffix:                     string
	binSize:                     number
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#SingleStatViewProperties: {
	type: "single-stat"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	prefix:            string
	tickPrefix:        string
	suffix:            string
	tickSuffix:        string
	staticLegend?:     #StaticLegend
	decimalPlaces:     #DecimalPlaces
	...
}
#HistogramViewProperties: {
	type: "histogram"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	xColumn:           string
	fillColumns: [...string]
	xDomain: [...number]
	xAxisLabel:                  string
	position:                    "overlaid" | "stacked"
	binCount:                    int
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#GaugeViewProperties: {
	type: "gauge"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	prefix:            string
	tickPrefix:        string
	suffix:            string
	tickSuffix:        string
	decimalPlaces:     #DecimalPlaces
	...
}
#TableViewProperties: {
	type: "table"
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	shape: "chronograf-v2"
	note:  string

	// If true, will display note when empty
	showNoteWhenEmpty: bool
	tableOptions: {
		// verticalTimeAxis describes the orientation of the table by
		// indicating whether the time axis will be displayed vertically
		verticalTimeAxis?: bool
		sortBy?:           #RenamableField

		// Wrapping describes the text wrapping style to be used in table
		// views
		wrapping?: "truncate" | "wrap" | "single-line"

		// fixFirstColumn indicates whether the first column of the table
		// should be locked
		fixFirstColumn?: bool
		...
	}

	// fieldOptions represent the fields retrieved by the query with
	// customization options
	fieldOptions: [...#RenamableField]

	// timeFormat describes the display format for time values
	// according to moment.js date formatting
	timeFormat:    string
	decimalPlaces: #DecimalPlaces
	...
}
#MarkdownViewProperties: {
	type:  "markdown"
	shape: "chronograf-v2"
	note:  string
	...
}
#CheckViewProperties: {
	type:    "check"
	shape:   "chronograf-v2"
	checkID: string
	check?:  #Check
	queries: [...#DashboardQuery]

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	legendColorizeRows?:         bool
	legendHide?:                 bool
	legendOpacity?:              number
	legendOrientationThreshold?: int
	...
}
#GeoViewLayer: #GeoCircleViewLayer | #GeoHeatMapViewLayer | #GeoPointMapViewLayer | #GeoTrackMapViewLayer
#GeoViewLayerProperties: {
	type: "heatmap" | "circleMap" | "pointMap" | "trackMap"
	...
}
#GeoCircleViewLayer: #GeoViewLayerProperties & {
	// Radius field
	radiusField:     string
	radiusDimension: #Axis

	// Circle color field
	colorField:     string
	colorDimension: #Axis

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]

	// Maximum radius size in pixels
	radius?: int

	// Interpolate circle color based on displayed value
	interpolateColors?: bool
	...
}
#GeoPointMapViewLayer: #GeoViewLayerProperties & {
	// Marker color field
	colorField:     string
	colorDimension: #Axis

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]

	// Cluster close markers together
	isClustered?: bool

	// An array for which columns to display in tooltip
	tooltipColumns?: [...string]
	...
}
#GeoTrackMapViewLayer: #GeoViewLayerProperties & {
	trackWidth:              _
	speed:                   _
	randomColors:            _
	trackPointVisualization: _
	...
} & {
	// Width of the track
	trackWidth?: int

	// Speed of the track animation
	speed?: int

	// Assign different colors to different tracks
	randomColors?: bool

	// Colors define color encoding of data into a visualization
	colors?: [...#DashboardColor]
	...
}
#GeoHeatMapViewLayer: #GeoViewLayerProperties & {
	// Intensity field
	intensityField:     string
	intensityDimension: #Axis

	// Radius size in pixels
	radius: int

	// Blur for heatmap points
	blur: int

	// Colors define color encoding of data into a visualization
	colors: [...#DashboardColor]
	...
}
#GeoViewProperties: {
	type: "geo"
	queries: [...#DashboardQuery]
	shape: "chronograf-v2"

	// Coordinates of the center of the map
	center: {
		// Latitude of the center of the map
		lat: number

		// Longitude of the center of the map
		lon: number
		...
	}

	// Zoom level used for initial display of the map
	zoom: >=1 & <=28

	// If true, map zoom and pan controls are enabled on the dashboard
	// view
	allowPanAndZoom: bool | *true

	// If true, search results get automatically regroupped so that
	// lon,lat and value are treated as columns
	detectCoordinateFields: bool | *true

	// If true, S2 column is used to calculate lat/lon
	useS2CellID?: bool

	// String to define the column
	s2Column?:      string
	latLonColumns?: #LatLonColumns

	// Define map type - regular, satellite etc.
	mapStyle?: string
	note:      string

	// If true, will display note when empty
	showNoteWhenEmpty: bool

	// Colors define color encoding of data into a visualization
	colors?: [...#DashboardColor]

	// List of individual layers shown in the map
	layers: [...#GeoViewLayer]
	...
}

// Object type to define lat/lon columns
#LatLonColumns: {
	lat: #LatLonColumn
	lon: #LatLonColumn
	...
}

// Object type for key and column definitions
#LatLonColumn: {
	// Key to determine whether the column is tag/field
	key: string

	// Column to look up Lat/Lon
	column: string
	...
}

// The viewport for a View's visualizations
#Axes: {
	x: #Axis
	y: #Axis
	...
}

// StaticLegend represents the options specific to the static
// legend
#StaticLegend: {
	colorizeRows?:         bool
	heightRatio?:          number
	show?:                 bool
	opacity?:              number
	orientationThreshold?: int
	valueAxis?:            string
	widthRatio?:           number
	...
}

// Indicates whether decimal places should be enforced, and how
// many digits it should show.
#DecimalPlaces: {
	// Indicates whether decimal point setting should be enforced
	isEnforced?: bool

	// The number of digits after decimal to display
	digits?: int
	...
}
#ConstantVariableProperties: null | bool | number | string | [...] | {
	type?: "constant"
	values?: [...string]
	...
}
#MapVariableProperties: null | bool | number | string | [...] | {
	type?: "map"
	values?: [string]: string
	...
}
#QueryVariableProperties: null | bool | number | string | [...] | {
	type?: "query"
	values?: {
		query?:    string
		language?: string
		...
	}
	...
}
#VariableProperties: #QueryVariableProperties | #ConstantVariableProperties | #MapVariableProperties
#ViewProperties:     #LinePlusSingleStatProperties | #XYViewProperties | #SingleStatViewProperties | #HistogramViewProperties | #GaugeViewProperties | #TableViewProperties | #MarkdownViewProperties | #CheckViewProperties | #ScatterViewProperties | #HeatmapViewProperties | #MosaicViewProperties | #BandViewProperties | #GeoViewProperties
#View:               null | bool | number | string | [...] | {
	links?: {
		self?: string
		...
	}
	id?:        string
	name:       string
	properties: #ViewProperties
	...
}
#Views: {
	links?: {
		self?: string
		...
	}
	views?: [...#View]
	...
}
#CellUpdate: {
	x?: int
	y?: int
	w?: int
	h?: int
	...
}
#CreateCell: {
	name?: string
	x?:    int
	y?:    int
	w?:    int
	h?:    int

	// Makes a copy of the provided view.
	usingView?: string
	...
}
#AnalyzeQueryResponse: {
	errors?: [...{
		line?:      int
		column?:    int
		character?: int
		message?:   string
		...
	}]
	...
}
#CellWithViewProperties: #Cell & {
	name?:       string
	properties?: #ViewProperties
	...
}
#Cell: {
	id?: string
	links?: {
		self?: string
		view?: string
		...
	}
	x?: int
	y?: int
	w?: int
	h?: int

	// The reference to a view from the views API.
	viewID?: string
	...
}
#CellsWithViewProperties: [...#CellWithViewProperties]
#Cells: [...#Cell]
#Secrets: null | bool | number | string | [...] | {
	[string]: string
}
#SecretKeys: {
	secrets?: [...string]
	...
}
#SecretKeysResponse: #SecretKeys & {
	links?: {
		self?: string
		org?:  string
		...
	}
	...
}
#CreateDashboardRequest: null | bool | number | string | [...] | {
	// The ID of the organization that owns the dashboard.
	orgID: string

	// The user-facing name of the dashboard.
	name: string

	// The user-facing description of the dashboard.
	description?: string
	...
}
#DashboardWithViewProperties: #CreateDashboardRequest & {
	links?: {
		self?:    #Link
		cells?:   #Link
		members?: #Link
		owners?:  #Link
		labels?:  #Link
		org?:     #Link
		...
	}
	id?: string
	meta?: {
		createdAt?: string
		updatedAt?: string
		...
	}
	cells?:  #CellsWithViewProperties
	labels?: #Labels
	...
}
#Dashboard: #CreateDashboardRequest & {
	links?: {
		self?:    #Link
		cells?:   #Link
		members?: #Link
		owners?:  #Link
		labels?:  #Link
		org?:     #Link
		...
	}
	id?: string
	meta?: {
		createdAt?: string
		updatedAt?: string
		...
	}
	cells?:  #Cells
	labels?: #Labels
	...
}
#Dashboards: {
	links?: #Links
	dashboards?: [...#Dashboard]
	...
}
#DocumentMeta: {
	name:         string
	type?:        string
	templateID?:  string
	description?: string
	version:      string
	createdAt?:   string
	updatedAt?:   string
	...
}
#Document: {
	id:   string
	meta: #DocumentMeta
	content: {
		...
	}
	labels?: #Labels
	links?: {
		// The document URL.
		self?: #Link
		...
	}
	...
}
#DocumentCreate: {
	meta: #DocumentMeta
	content: {
		...
	}

	// The organization Name. Specify either `orgID` or `org`.
	org?: string

	// The organization Name. Specify either `orgID` or `org`.
	orgID?: string

	// An array of label IDs to be added as labels to the document.
	labels?: [...string]
	...
}
#DocumentUpdate: {
	meta?: #DocumentMeta
	content?: {
		...
	}
	...
}
#DocumentListEntry: {
	id:      string
	meta:    #DocumentMeta
	labels?: #Labels
	links?: {
		// The document URL.
		self?: #Link
		...
	}
	...
}
#Documents: {
	documents?: [...#DocumentListEntry]
	...
}
#TelegrafRequest: {
	name?:        string
	description?: string
	metadata?: {
		buckets?: [...string]
		...
	}
	config?: string
	orgID?:  string
	...
}
#Telegraf: #TelegrafRequest & {
	id?: string
	links?: {
		self?:    #Link
		labels?:  #Link
		members?: #Link
		owners?:  #Link
		...
	}
	labels?: #Labels
	...
}
#Telegrafs: {
	configurations?: [...#Telegraf]
	...
}
#TelegrafPlugin: {
	type?:        string
	name?:        string
	description?: string
	config?:      string
	...
}
#TelegrafPlugins: {
	version?: string
	os?:      string
	plugins?: [...#TelegrafPlugin]
	...
}
#IsOnboarding: {
	// True means that the influxdb instance has NOT had initial
	// setup; false means that the database has been setup.
	allowed?: bool
	...
}
#PasswordResetBody: null | bool | number | string | [...] | {
	password: string
	...
}
#AddResourceMemberRequestBody: {
	id:    string
	name?: string
	...
}
#Ready: {
	status?:  "ready"
	started?: string
	up?:      string
	...
}
#HealthCheck: {
	name:     string
	message?: string
	checks?: [...#HealthCheck]
	status:   "pass" | "fail"
	version?: string
	commit?:  string
	...
}
#Labels: [...#Label]
#Label: {
	id?:    string
	orgID?: string
	name?:  string

	// Key/Value pairs associated with this label. Keys can be removed
	// by sending an update with an empty value.
	properties?: {
		[string]: string
	}
	...
}
#LabelCreateRequest: {
	orgID: string
	name:  string

	// Key/Value pairs associated with this label. Keys can be removed
	// by sending an update with an empty value.
	properties?: {
		[string]: string
	}
	...
}
#LabelUpdate: {
	name?: string

	// Key/Value pairs associated with this label. Keys can be removed
	// by sending an update with an empty value.
	properties?: {
		[string]: string
	}
	...
}
#LabelMapping: {
	labelID?: string
	...
}
#LabelsResponse: {
	labels?: #Labels
	links?:  #Links
	...
}
#LabelResponse: {
	label?: #Label
	links?: #Links
	...
}

// Contains the AST for the supplied Flux query
#ASTResponse: {
	ast?: #Package
	...
}
#WritePrecision: "ms" | "s" | "us" | "ns"
#TaskCreateRequest: {
	// The ID of the organization that owns this Task.
	orgID?: string

	// The name of the organization that owns this Task.
	org?:    string
	status?: #TaskStatusType

	// The Flux script to run for this task.
	flux: string

	// An optional description of the task.
	description?: string
	...
}
#TaskUpdateRequest: {
	status?: #TaskStatusType

	// The Flux script to run for this task.
	flux?: string

	// Override the 'name' option in the flux script.
	name?: string

	// Override the 'every' option in the flux script.
	every?: string

	// Override the 'cron' option in the flux script.
	cron?: string

	// Override the 'offset' option in the flux script.
	offset?: string

	// An optional description of the task.
	description?: string
	...
}

// Rendered flux that backs the check or notification.
#FluxResponse: null | bool | number | string | [...] | {
	flux?: string
	...
}
#CheckPatch: {
	name?:        string
	description?: string
	status?:      "active" | "inactive"
	...
}
#CheckDiscriminator: #DeadmanCheck | #ThresholdCheck | #CustomCheck
#Check:              #CheckDiscriminator
#PostCheck:          #CheckDiscriminator
#Checks:             null | bool | number | string | [...] | {
	checks?: [...#Check]
	links?: #Links
	...
}
#CheckBase: null | bool | number | string | [...] | {
	id?:  string
	name: string

	// The ID of the organization that owns this check.
	orgID: string

	// The ID of the task associated with this check.
	taskID?: string

	// The ID of creator used to create this check.
	ownerID?:   string
	createdAt?: string
	updatedAt?: string
	query:      #DashboardQuery
	status?:    #TaskStatusType

	// An optional description of the check.
	description?: string

	// Timestamp of latest scheduled, completed run, RFC3339.
	latestCompleted?: string
	lastRunStatus?:   "failed" | "success" | "canceled"
	lastRunError?:    string
	labels?:          #Labels
	links?: {
		// URL for this check
		self?: #Link

		// URL to retrieve labels for this check
		labels?: #Link

		// URL to retrieve members for this check
		members?: #Link

		// URL to retrieve owners for this check
		owners?: #Link

		// URL to retrieve flux script for this check
		query?: #Link
		...
	}
	...
}
#ThresholdCheck: #CheckBase & {
	type: "threshold"
	thresholds?: [...#Threshold]

	// Check repetition interval.
	every?: string

	// Duration to delay after the schedule, before executing check.
	offset?: string

	// List of tags to write to each status.
	tags?: [...{
		key?:   string
		value?: string
		...
	}]

	// The template used to generate and write a status message.
	statusMessageTemplate?: string
	...
}
#Threshold:    #GreaterThreshold | #LesserThreshold | #RangeThreshold
#DeadmanCheck: #CheckBase & {
	type: "deadman"

	// String duration before deadman triggers.
	timeSince?: string

	// String duration for time that a series is considered stale and
	// should not trigger deadman.
	staleTime?: string

	// If only zero values reported since time, trigger an alert
	reportZero?: bool
	level?:      #CheckStatusLevel

	// Check repetition interval.
	every?: string

	// Duration to delay after the schedule, before executing check.
	offset?: string

	// List of tags to write to each status.
	tags?: [...{
		key?:   string
		value?: string
		...
	}]

	// The template used to generate and write a status message.
	statusMessageTemplate?: string
	...
}
#CustomCheck: #CheckBase & {
	type: "custom"
	...
}
#ThresholdBase: null | bool | number | string | [...] | {
	level?: #CheckStatusLevel

	// If true, only alert if all values meet threshold.
	allValues?: bool
	...
}
#GreaterThreshold: #ThresholdBase & {
	type:  "greater"
	value: number
	...
}
#LesserThreshold: #ThresholdBase & {
	type:  "lesser"
	value: number
	...
}
#RangeThreshold: #ThresholdBase & {
	type:   "range"
	min:    number
	max:    number
	within: bool
	...
}

// The state to record if check matches a criteria.
#CheckStatusLevel: "UNKNOWN" | "OK" | "INFO" | "CRIT" | "WARN"

// The state to record if check matches a criteria.
#RuleStatusLevel: "UNKNOWN" | "OK" | "INFO" | "CRIT" | "WARN" | "ANY"
#NotificationRuleUpdate: {
	name?:        string
	description?: string
	status?:      "active" | "inactive"
	...
}
#NotificationRuleDiscriminator: #SlackNotificationRule | #SMTPNotificationRule | #PagerDutyNotificationRule | #HTTPNotificationRule | #TelegramNotificationRule
#NotificationRule:              #NotificationRuleDiscriminator
#PostNotificationRule:          #NotificationRuleDiscriminator
#NotificationRules:             null | bool | number | string | [...] | {
	notificationRules?: [...#NotificationRule]
	links?: #Links
	...
}
#NotificationRuleBase: {
	// Timestamp of latest scheduled, completed run, RFC3339.
	latestCompleted?: string
	lastRunStatus?:   "failed" | "success" | "canceled"
	lastRunError?:    string
	id?:              string
	endpointID:       string

	// The ID of the organization that owns this notification rule.
	orgID: string

	// The ID of the task associated with this notification rule.
	taskID?: string

	// The ID of creator used to create this notification rule.
	ownerID?:   string
	createdAt?: string
	updatedAt?: string
	status:     #TaskStatusType

	// Human-readable name describing the notification rule.
	name:        string
	sleepUntil?: string

	// The notification repetition interval.
	every?: string

	// Duration to delay after the schedule, before executing check.
	offset?:      string
	runbookLink?: string

	// Don't notify me more than <limit> times every <limitEvery>
	// seconds. If set, limit cannot be empty.
	limitEvery?: int

	// Don't notify me more than <limit> times every <limitEvery>
	// seconds. If set, limitEvery cannot be empty.
	limit?: int

	// List of tag rules the notification rule attempts to match.
	tagRules?: [...#TagRule]

	// An optional description of the notification rule.
	description?: string

	// List of status rules the notification rule attempts to match.
	statusRules: [_, ...] & [...#StatusRule]
	labels?:     #Labels
	links?: {
		// URL for this endpoint.
		self?: #Link

		// URL to retrieve labels for this notification rule.
		labels?: #Link

		// URL to retrieve members for this notification rule.
		members?: #Link

		// URL to retrieve owners for this notification rule.
		owners?: #Link

		// URL to retrieve flux script for this notification rule.
		query?: #Link
		...
	}
	...
}
#TagRule: {
	key?:      string
	value?:    string
	operator?: "equal" | "notequal" | "equalregex" | "notequalregex"
	...
}
#StatusRule: {
	currentLevel?:  #RuleStatusLevel
	previousLevel?: #RuleStatusLevel
	count?:         int
	period?:        string
	...
}
#HTTPNotificationRuleBase: {
	type: "http"
	url?: string
	...
}
#HTTPNotificationRule: #NotificationRuleBase & #HTTPNotificationRuleBase
#SlackNotificationRuleBase: {
	type:            "slack"
	channel?:        string
	messageTemplate: string
	...
}
#SlackNotificationRule: #NotificationRuleBase & #SlackNotificationRuleBase
#SMTPNotificationRule:  #NotificationRuleBase & #SMTPNotificationRuleBase
#SMTPNotificationRuleBase: {
	type:            "smtp"
	subjectTemplate: string
	bodyTemplate?:   string
	to:              string
	...
}
#PagerDutyNotificationRule: #NotificationRuleBase & #PagerDutyNotificationRuleBase
#PagerDutyNotificationRuleBase: {
	type:            "pagerduty"
	messageTemplate: string
	...
}
#TelegramNotificationRule: #NotificationRuleBase & #TelegramNotificationRuleBase
#TelegramNotificationRuleBase: {
	// The discriminator between other types of notification rules is
	// "telegram".
	type: "telegram"

	// The message template as a flux interpolated string.
	messageTemplate: string

	// Parse mode of the message text per
	// https://core.telegram.org/bots/api#formatting-options .
	// Defaults to "MarkdownV2" .
	parseMode?: "MarkdownV2" | "HTML" | "Markdown"

	// Disables preview of web links in the sent messages when "true".
	// Defaults to "false" .
	disableWebPagePreview?: bool
	channel:                _
	...
}
#NotificationEndpointUpdate: {
	name?:        string
	description?: string
	status?:      "active" | "inactive"
	...
}
#NotificationEndpointDiscriminator: #SlackNotificationEndpoint | #PagerDutyNotificationEndpoint | #HTTPNotificationEndpoint | #TelegramNotificationEndpoint
#NotificationEndpoint:              #NotificationEndpointDiscriminator
#PostNotificationEndpoint:          #NotificationEndpointDiscriminator
#NotificationEndpoints:             null | bool | number | string | [...] | {
	notificationEndpoints?: [...#NotificationEndpoint]
	links?: #Links
	...
}
#NotificationEndpointBase: {
	id?:        string
	orgID?:     string
	userID?:    string
	createdAt?: string
	updatedAt?: string

	// An optional description of the notification endpoint.
	description?: string
	name:         string

	// The status of the endpoint.
	status?: "active" | "inactive" | *"active"
	labels?: #Labels
	links?: {
		// URL for this endpoint.
		self?: #Link

		// URL to retrieve labels for this endpoint.
		labels?: #Link

		// URL to retrieve members for this endpoint.
		members?: #Link

		// URL to retrieve owners for this endpoint.
		owners?: #Link
		...
	}
	type: #NotificationEndpointType
	...
}
#SlackNotificationEndpoint: #NotificationEndpointBase & {
	// Specifies the URL of the Slack endpoint. Specify either `URL`
	// or `Token`.
	url?: string

	// Specifies the API token string. Specify either `URL` or
	// `Token`.
	token?: string
	...
}
#PagerDutyNotificationEndpoint: #NotificationEndpointBase & {
	clientURL?: string
	routingKey: string
	...
}
#HTTPNotificationEndpoint: #NotificationEndpointBase & {
	url:              string
	username?:        string
	password?:        string
	token?:           string
	method:           "POST" | "GET" | "PUT"
	authMethod:       "none" | "basic" | "bearer"
	contentTemplate?: string

	// Customized headers.
	headers?: {
		[string]: string
	}
	...
}
#TelegramNotificationEndpoint: #NotificationEndpointBase & {
	// Specifies the Telegram bot token. See
	// https://core.telegram.org/bots#creating-a-new-bot .
	token: string

	// ID of the telegram channel, a chat_id in
	// https://core.telegram.org/bots/api#sendmessage .
	channel: string
	...
}
#NotificationEndpointType: "slack" | "pagerduty" | "http" | "telegram"
#DBRP: {
	// the mapping identifier
	id: string

	// the organization ID that owns this mapping.
	orgID: string

	// the bucket ID used as target for the translation.
	bucketID: string

	// InfluxDB v1 database
	database: string

	// InfluxDB v1 retention policy
	retention_policy: string

	// Specify if this mapping represents the default retention policy
	// for the database specificed.
	default: bool
	links?:  #Links
	...
}
#DBRPs: null | bool | number | string | [...] | {
	content?: [...#DBRP]
	...
}
#DBRPUpdate: null | bool | number | string | [...] | {
	// InfluxDB v1 retention policy
	retention_policy?: string
	default?:          bool
	...
}
#DBRPCreate: {
	// the organization ID that owns this mapping.
	orgID?: string

	// the organization that owns this mapping.
	org?: string

	// the bucket ID used as target for the translation.
	bucketID: string

	// InfluxDB v1 database
	database: string

	// InfluxDB v1 retention policy
	retention_policy: string

	// Specify if this mapping represents the default retention policy
	// for the database specificed.
	default?: bool
	...
}
#DBRPGet: {
	content: #DBRP
	...
}
#SchemaType:    "implicit" | "explicit"
#Authorization: #AuthorizationUpdateRequest & {
	createdAt?: string
	updatedAt?: string

	// ID of org that authorization is scoped to.
	orgID?: string

	// List of permissions for an auth. An auth must have at least one
	// Permission.
	permissions?: [_, ...] & [...#Permission]
	id?:          string

	// Passed via the Authorization Header and Token Authentication
	// type.
	token?: string

	// ID of user that created and owns the token.
	userID?: string

	// Name of user that created and owns the token.
	user?: string

	// Name of the org token is scoped to.
	org?: string
	links?: {
		self?: #Link
		user?: #Link
		...
	}
	...
} & {
	orgID:       _
	permissions: _
	...
}
#AuthorizationPostRequest: #AuthorizationUpdateRequest & {
	// ID of org that authorization is scoped to.
	orgID?: string

	// ID of user that authorization is scoped to.
	userID?: string

	// List of permissions for an auth. An auth must have at least one
	// Permission.
	permissions?: [_, ...] & [...#Permission]
	...
} & {
	orgID:       _
	permissions: _
	...
}
#LegacyAuthorizationPostRequest: #AuthorizationUpdateRequest & {
	// ID of org that authorization is scoped to.
	orgID?: string

	// ID of user that authorization is scoped to.
	userID?: string

	// Token (name) of the authorization
	token?: string

	// List of permissions for an auth. An auth must have at least one
	// Permission.
	permissions?: [_, ...] & [...#Permission]
	...
} & {
	orgID:       _
	permissions: _
	...
}
#Authorizations: {
	links?: #Links
	authorizations?: [...#Authorization]
	...
}
#Permission: null | bool | number | string | [...] | {
	action:   "read" | "write"
	resource: #Resource
	...
}
#Resource: {
	type: "authorizations" | "buckets" | "dashboards" | "orgs" | "sources" | "tasks" | "telegrafs" | "users" | "variables" | "scrapers" | "secrets" | "labels" | "views" | "documents" | "notificationRules" | "notificationEndpoints" | "checks" | "dbrp" | "notebooks"

	// If ID is set that is a permission for a specific resource. if
	// it is not set it is a permission for all resources of that
	// resource type.
	id?: string

	// Optional name of the resource if the resource has a name field.
	name?: string

	// If orgID is set that is a permission for all resources owned my
	// that org. if it is not set it is a permission for all
	// resources of that resource type.
	orgID?: string

	// Optional name of the organization of the organization with
	// orgID.
	org?: string
	...
}
#User: null | bool | number | string | [...] | {
	id?:      string
	oauthID?: string
	name:     string

	// If inactive the user is inactive.
	status?: "active" | "inactive" | *"active"
	...
}
#Users: {
	links?: {
		self?: string
		...
	}
	users?: [...#UserResponse]
	...
}
#OnboardingRequest: {
	username:                string
	password?:               string
	org:                     string
	bucket:                  string
	retentionPeriodSeconds?: int

	// Retention period *in nanoseconds* for the new bucket. This
	// key's name has been misleading since OSS 2.0 GA, please
	// transition to use `retentionPeriodSeconds`
	retentionPeriodHrs?: int @deprecated()

	// Authentication token to set on the initial user. If not
	// specified, the server will generate a token.
	token?: string
	...
}
#OnboardingResponse: {
	user?:   #UserResponse
	org?:    #Organization
	bucket?: #Bucket
	auth?:   #Authorization
	...
}
#Variable: {
	links?: {
		self?:   string
		org?:    string
		labels?: string
		...
	}
	id?:          string
	orgID:        string
	name:         string
	description?: string
	selected?: [...string]
	labels?:    #Labels
	arguments:  #VariableProperties
	createdAt?: string
	updatedAt?: string
	...
}
#Variables: {
	variables?: [...#Variable]
	...
}
#Source: {
	links?: {
		self?:    string
		query?:   string
		health?:  string
		buckets?: string
		...
	}
	id?:                 string
	orgID?:              string
	default?:            bool
	name?:               string
	type?:               "v1" | "v2" | "self"
	url?:                string
	insecureSkipVerify?: bool
	telegraf?:           string
	token?:              string
	username?:           string
	password?:           string
	sharedSecret?:       string
	metaUrl?:            string
	defaultRP?:          string
	languages?: [..."flux" | "influxql"]
	...
}
#Sources: {
	links?: {
		self?: string
		...
	}
	sources?: [...#Source]
	...
}
#ScraperTargetRequest: {
	// The name of the scraper target.
	name?: string

	// The type of the metrics to be parsed.
	type?: "prometheus"

	// The URL of the metrics endpoint.
	url?: string

	// The organization ID.
	orgID?: string

	// The ID of the bucket to write to.
	bucketID?: string

	// Skip TLS verification on endpoint.
	allowInsecure?: bool | *false
	...
}
#ScraperTargetResponse: #ScraperTargetRequest & {
	id?: string

	// The name of the organization.
	org?: string

	// The bucket name.
	bucket?: string
	links?: {
		self?:         #Link
		members?:      #Link
		owners?:       #Link
		bucket?:       #Link
		organization?: #Link
		...
	}
	...
}
#ScraperTargetResponses: {
	configurations?: [...#ScraperTargetResponse]
	...
}
#MetadataBackup: {
	kv:      string
	sql:     string
	buckets: #BucketMetadataManifests
	...
}
#BucketMetadataManifests: [...#BucketMetadataManifest]
#BucketMetadataManifest: {
	organizationID:         string
	organizationName:       string
	bucketID:               string
	bucketName:             string
	description?:           string
	defaultRetentionPolicy: string
	retentionPolicies:      #RetentionPolicyManifests
	...
}
#RetentionPolicyManifests: [...#RetentionPolicyManifest]
#RetentionPolicyManifest: {
	name:               string
	replicaN:           int
	duration:           int
	shardGroupDuration: int
	shardGroups:        #ShardGroupManifests
	subscriptions:      #SubscriptionManifests
	...
}
#ShardGroupManifests: [...#ShardGroupManifest]
#ShardGroupManifest: {
	id:           int
	startTime:    string
	endTime:      string
	deletedAt?:   string
	truncatedAt?: string
	shards:       #ShardManifests
	...
}
#ShardManifests: [...#ShardManifest]
#ShardManifest: {
	id:          int
	shardOwners: #ShardOwners
	...
}
#ShardOwners: [...#ShardOwner]
#ShardOwner: {
	// ID of the node that owns a shard.
	nodeID: int
	...
}
#SubscriptionManifests: [...#SubscriptionManifest]
#SubscriptionManifest: {
	name: string
	mode: string
	destinations: [...string]
	...
}
#RestoredBucketMappings: {
	// New ID of the restored bucket
	id:            string
	name:          string
	shardMappings: #BucketShardMappings
	...
}
#BucketShardMappings: [...#BucketShardMapping]
#BucketShardMapping: {
	oldId: int
	newId: int
	...
}
