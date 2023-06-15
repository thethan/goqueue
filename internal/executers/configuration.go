package executers

type Configuration struct {
	Sprintf     string        `json:"sprintf" yaml:"sprintf"`
	SprintfArgs []interface{} `json:"sprintfArgs" yaml:"sprintfArgs"`
	Command     []string      `json:"command" yaml:"command"`
	Args        []interface{} `json:"args" yaml:"args"`

	StdErr []byte `json:"stdErr" yaml:"stdErr"`
	StdOut []byte `json:"stdOut" yaml:"stdOut"`
}
