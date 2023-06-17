package internal

import "net/http"

// [string key: any value] map type.
type StringKeyMap = map[string]interface{}

// [string key: string value] map type.
type StringMap = map[string]string

// Type for payload - both for its template specified config JSON and for
// actual resulting payload (which is then converted to JSON).
type PayloadType = StringKeyMap

// Type for a container of variables for expression evaluator.
type EvalVariables = StringMap

// Function type for passing callback into binaryOpEval() function.
type binaryOpEvalCallback = func(string, string, string) (string, error)

// Struct representing global Reporter's runtime settings and stuff.
type ReporterSettings struct {
	SelfDir        string // Directory path where the compiled reporter binary is.
	ConfigJsonPath string // Populated via CLI argument "--config", if set.
	JustTry        bool   // Populated via CLI argument "--try", if set.
	VerboseMode    bool   // Populated via CLI argument "--verbose", if set.
	DaemonMode     bool   // True by default, false if "--try"
	ForegroundMode bool   // False by default, true if "--foreground"
}

// Struct representing config read from config.json file.
type Config struct {
	Target    []string
	Gatherers []string
	Env       map[string]string
	Payload   PayloadType
}

type Reporter struct {
	ConfigJson Config
	HttpClient *http.Client
}

// Struct that allows us to wrap some gatherer's result's with information about
// the gatherer's original order of execution (within a single gather-loop).
type OrderedGathererResult struct {
	index     int
	gatherer  string
	exitError error
	data      StringMap
}
