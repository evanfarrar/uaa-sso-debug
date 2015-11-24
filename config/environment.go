package config

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/ryanmoran/viron"
)

type Environment struct {
	Domain              string
	EncryptionKey       string `env:"ENCRYPTION_KEY"        env-required:"true"`
	Host                string `env:"HOST"                  env-default:"localhost"`
	LoginHost           string `env:"LOGIN_HOST"            env-required:"true"`
	Port                string `env:"PORT"                  env-default:"3000"`
	Scheme              string `env:"SCHEME"                env-required:"true"`
	UAAClientID         string `env:"UAA_CLIENT_ID"         env-required:"true"`
	LogFile             string `env:"LOG_FILE"`
	UAAClientSecret     string `env:"UAA_CLIENT_SECRET"     env-required:"true"`
	UAAHost             string `env:"UAA_HOST"              env-required:"true"`
	VerifySSL           bool   `env:"VERIFY_SSL"            env-default:"true"`
}

func NewEnvironment() Environment {
	env := Environment{}
	err := viron.Parse(&env)
	if err != nil {
		panic(err)
	}

	env.UAAHost = env.ApplySchemeToHost(env.UAAHost)
	env.LoginHost = env.ApplySchemeToHost(env.LoginHost)
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
