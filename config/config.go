package config

//ProgramConfig struct exposes the parsed program configuration
type ProgramConfig struct {
	Gcs gcs
	Web web
}
type gcs struct {
	ServiceAccountFilePath string
	Buckets                map[string]string
}
type web struct {
	Port  int16
	OAuth oauth
}
type creds struct {
	Username string
	Password string
}

type oauth struct {
	ClientID           string
	ClientSecret       string
	AllowedHostDomains []string
	SessionSecret      string
}
