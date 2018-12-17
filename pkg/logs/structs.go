package logs

const (
	// DependencyStorage identifies a storage failure
	DependencyStorage = "storage"

	//DependencyQueuer identifies a queuer failure
	DependencyQueuer = "queuer"

	// DependencyMarker identifies a marker failure
	DependencyMarker = "marker"

	// DependencyDigester identifies a digester failure
	DependencyDigester = "digester"

	// DependencyGrapher identifies a grapher failure
	DependencyGrapher = "grapher"
)

// DependencyFailure is logged when a downstream dependency fails
type DependencyFailure struct {
	Dependency string `logevent:"dependency"`
	Reason     string `logevent:"reason"`
	Message    string `logevent:"message,default=dependency-failure"`
}

// UnknownFailure is logged when an unexpected error occurs
type UnknownFailure struct {
	Reason  string `logevent:"reason"`
	Message string `logevent:"message,default=unknown-failure"`
}

// InvalidInput is logged when the input provided is not valid
type InvalidInput struct {
	Reason  string `logevent:"reason"`
	Message string `logevent:"message,default=invalid-input"`
}

// NotFound is logged when the requested resource is not found
type NotFound struct {
	Reason  string `logevent:"reason"`
	Message string `logevent:"message,default=not-found"`
}

// Conflict is logged when the input provided is not valid
type Conflict struct {
	Reason  string `logevent:"reason"`
	Message string `logevent:"message,default=conflict"`
}
