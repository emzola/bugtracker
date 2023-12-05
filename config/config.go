package config

// AppConfiguration defines configuration values. Values are read via
// command-line flags and environment variables.
type AppConfiguration struct {
	Port     int
	Env      string
	Database struct {
		Dsn          string
		MaxOpenConns int
		MaxIdleConns int
		MaxIdleTime  string
	}
	Smtp struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}
	Jwt struct {
		Secret string
	}
	Limiter struct {
		Rps     float64
		Burst   int
		Enabled bool
	}
	Cors struct {
		TrustedOrigins []string
	}
}
