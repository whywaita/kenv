package extractor

type EnvVar struct {
	Name      string
	Value     string
	Source    EnvVarSource
	IsSecret  bool
	SecretRef *SecretKeyRef
	ConfigRef *ConfigMapKeyRef
}

type EnvVarSource int

const (
	SourceDirect EnvVarSource = iota
	SourceSecret
	SourceConfigMap
)

type SecretKeyRef struct {
	Name string
	Key  string
}

type ConfigMapKeyRef struct {
	Name string
	Key  string
}

type Options struct {
	Container string
}