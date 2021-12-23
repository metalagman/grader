package aws

type Config struct {
	Region string `mapstructure:"region"`
	Key    string `mapstructure:"access_key_id"`
	Secret string `mapstructure:"secret_access_key"`
	Bucket string `mapstructure:"bucket"`
}
