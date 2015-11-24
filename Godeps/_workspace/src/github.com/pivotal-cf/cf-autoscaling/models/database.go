package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/go-gorp/gorp"
	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/log"

	_ "github.com/go-sql-driver/mysql"
)

var _database *DB
var mutex sync.Mutex

const DatabaseInstanceName = "autoscale-mysql"

type DB struct {
	Uri        string
	Connection *gorp.DbMap
}

type ServiceRepresentation struct {
	Name        string `json:"name"`
	Credentials struct {
		URI string `json:"uri"`
	} `json:"credentials"`
}

type ServicesEnvVar map[string][]ServiceRepresentation

func NewDB() *DB {
	env := config.NewEnvironment()

	return &DB{
		Uri: formatUri(env.DatabaseURL),
	}
}

func Database() *DB {
	defer mutex.Unlock()
	mutex.Lock()
	if _database == nil || _database.Connection == nil {
		_database = NewDB()

		env := config.NewEnvironment()
		if _database.Uri != "" {
			_database.connect()
			_database.enableLogging(env)
			_database.mapTables()
			_database.migrate(env)
		}
	}
	return _database
}

func formatUri(uri string) string {
	parsed, err := url.Parse(uri)
	if err != nil {
		panic(err)
	}

	password, _ := parsed.User.Password()
	return fmt.Sprintf("%s:%s@tcp(%s)%s?parseTime=true&timeout=10s", parsed.User.Username(), password, parsed.Host, parsed.Path)
}

func (d *DB) Disconnect() {
	d.Connection.Db.Close()
}

func (d *DB) Reconnect() {
	env := config.NewEnvironment()
	d.connect()
	d.enableLogging(env)
	d.mapTables()
}

func (d *DB) setEnv(env config.Environment) {
	servicesVar := make(ServicesEnvVar, 0)
	json.Unmarshal([]byte(env.VCAPServices), &servicesVar)

	for _, serviceRepresentations := range servicesVar {
		for _, serviceRepresentation := range serviceRepresentations {
			if serviceRepresentation.Name == DatabaseInstanceName {
				d.Uri = formatUri(serviceRepresentation.Credentials.URI)
				return
			}
		}
	}
}

func (d *DB) enableLogging(env config.Environment) {
	if env.DBLoggingEnabled {
		d.Connection.TraceOn("[DB]", log.Logger)
	}
}

func (d *DB) connect() {
	if d.Uri == "" {
		err := errors.New("No Database URI defined.")
		if err != nil {
			panic(err)
		}
	}

	db, err := sql.Open("mysql", d.Uri)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	d.Connection = &gorp.DbMap{
		Db: db,
		Dialect: gorp.MySQLDialect{
			Engine:   "InnoDB",
			Encoding: "UTF8",
		},
	}
}

func (d *DB) migrate(env config.Environment) {
	dbDriver := goose.DBDriver{
		Name:    "mysql",
		OpenStr: d.Uri,
		Import:  "github.com/go-sql-driver/mysql",
		Dialect: &goose.MySqlDialect{},
	}

	migrationsDir := env.Root + "/models/migrations"

	dbConf := goose.DBConf{
		MigrationsDir: migrationsDir,
		Env:           "autoscale",
		Driver:        dbDriver,
	}

	current, err := goose.GetDBVersion(&dbConf)
	if err != nil {
		panic(err)
	}

	target, err := goose.GetMostRecentDBVersion(migrationsDir)
	if err != nil {
		panic(err)
	}

	if current != target {
		fmt.Println("Running migrations...")
		err = goose.RunMigrations(&dbConf, migrationsDir, target)
		if err != nil {
			panic(err)
		}
	}
}

func (d *DB) mapTables() {
	conn := d.Connection
	conn.AddTableWithName(ServiceInstance{}, "service_instances").SetKeys(false, "Guid")
	conn.AddTableWithName(ServiceBinding{}, "service_bindings").SetKeys(false, "Guid")
	conn.AddTableWithName(Reading{}, "readings").SetKeys(true, "ID")
	conn.AddTableWithName(ScalingDecision{}, "scaling_decisions").SetKeys(true, "ID")
	conn.AddTableWithName(ScheduledRule{}, "scheduled_rules").SetKeys(true, "ID")
	conn.AddTableWithName(kv{}, "key_value_store").SetKeys(false, "key")
}
