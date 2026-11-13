package config

type Config struct{}

func ProcessEnvironmentVariables() (*Config, error) {
	// In all cases the default behavior should be for the docker compose setup
	env := Config{}

	return &env, nil
}
