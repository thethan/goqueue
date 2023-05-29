package job

type RawJobType string

const (
	JsonRawJobType  RawJobType = "json"
	ProtoRawJobType RawJobType = "proto"

	YamlRawJobType RawJobType = "yaml"
)

type Field struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Default string `json:"default" yaml:"default"`
}
type Configuration struct {
	Type RawJobType `json:"type" yaml:"type"`
}
