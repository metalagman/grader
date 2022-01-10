package aws

type Config struct {
	Endpoint   string `mapstructure:"endpoint"`
	Region     string `mapstructure:"region"`
	Key        string `mapstructure:"key"`
	Secret     string `mapstructure:"secret"`
	Bucket     string `mapstructure:"bucket"`
	DisableSSL bool   `mapstructure:"disable_ssl"`
}
