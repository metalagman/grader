package aws

type Config struct {
	Endpoint string `mapstructure:"endpoint"`
	Region   string `mapstructure:"region"`
	Key      string `mapstructure:"access_key_id"`
	Secret   string `mapstructure:"secret_access_key"`
	Bucket   string `mapstructure:"bucket"`
}
