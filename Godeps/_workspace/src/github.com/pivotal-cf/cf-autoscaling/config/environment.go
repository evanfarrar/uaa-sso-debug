package config

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryanmoran/viron"
)

type Environment struct {
	BasicAuthPassword   string `env:"BASIC_AUTH_PASSWORD"   env-required:"true"`
	BasicAuthUsername   string `env:"BASIC_AUTH_USERNAME"   env-required:"true"`
	CCHost              string `env:"CC_HOST"               env-required:"true"`
	DatabaseURL         string `env:"DATABASE_URL"`
	DBLoggingEnabled    bool   `env:"DB_LOGGING_ENABLED"`
	Domain              string
	EncryptionKey       string `env:"ENCRYPTION_KEY"        env-required:"true"`
	HTTPLoggingEnabled  bool   `env:"HTTP_LOGGING_ENABLED"`
	Host                string `env:"HOST"                  env-default:"localhost"`
	LogFile             string `env:"LOG_FILE"`
	LoginHost           string `env:"LOGIN_HOST"            env-required:"true"`
	MetricsPollInterval int    `env:"METRICS_POLL_INTERVAL" env-default:"10"`
	NotificationsHost   string `env:"NOTIFICATIONS_HOST"`
	Port                string `env:"PORT"                  env-default:"3000"`
	PublicPath          string
	Root                string `env:"ROOT"                  env-required:"true"`
	Scheme              string `env:"SCHEME"                env-required:"true"`
	UAAClientID         string `env:"UAA_CLIENT_ID"         env-required:"true"`
	UAAClientSecret     string `env:"UAA_CLIENT_SECRET"     env-required:"true"`
	UAAHost             string `env:"UAA_HOST"              env-required:"true"`
	VCAPServices        string `env:"VCAP_SERVICES"`
	VerifySSL           bool   `env:"VERIFY_SSL"            env-default:"true"`

	VCAPApplication struct {
		InstanceIndex int `json:"instance_index"`
	} `env:"VCAP_APPLICATION"`
}

func NewEnvironment() Environment {
	env := Environment{}
	err := viron.Parse(&env)
	if err != nil {
		panic(err)
	}

	env.Root = env.ExpandRoot(env.Root)
	env.UAAHost = env.ApplySchemeToHost(env.UAAHost)
	env.CCHost = env.ApplySchemeToHost(env.CCHost)
	env.LoginHost = env.ApplySchemeToHost(env.LoginHost)
	env.NotificationsHost = env.ApplySchemeToHost(env.NotificationsHost)
	env.PublicPath = filepath.Clean(env.Root + "/public")
	env.Domain = env.BuildDomain()

	return env
}

func (env Environment) ExpandRoot(root string) string {
	return os.ExpandEnv(root)
}

func (env Environment) ApplySchemeToHost(host string) string {
	if host == "" {
		return ""
	}

	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	return fmt.Sprintf("%s://%s", env.Scheme, host)
}

func (env Environment) BuildDomain() string {
	port := env.Port
	if env.Host != "localhost" {
		port = ""
	}

	return fmt.Sprintf("%s://%s", env.Scheme, strings.TrimSuffix(net.JoinHostPort(env.Host, port), ":"))
}
