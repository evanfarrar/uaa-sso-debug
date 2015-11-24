#cf-autoscaling
[![Build Status](https://magnum.travis-ci.com/evanfarrar/uaa-sso-debug.svg?token=X3mC6HYvEdT5EgNCXnXj&branch=master)](https://magnum.travis-ci.com/evanfarrar/uaa-sso-debug)

##Setup
The easiest way to get a running version of the service is to deploy it with the [bosh release](https://github.com/evanfarrar/uaa-sso-debug-release).

Developers can run the service locally by setting up a MySQL database and configuring the development [environment variables](#environment-variables) to point to that instance. Additionally, a UAA client is required to handle permissions for the service and its users. Instructions to setup the UAA client can be found [here](#uaa-setup)

Steps to run locally are as follows:

1. clone repo into `$GOPATH/src/github.com/pivotal-cf`
2. `go get github.com/kr/godep`
3. Make sure `$GOPATH/bin` is in `$PATH`
4. Run `godep restore`
5. Ensure `npm` is installed
6. `cd ./assets` and run `npm install`
7. Ensure there is a `bin/env/test` file that exports the environment variables shown in `bin/env/example`
8. Run the test suite with `bin/test`
9. Run the server with `bin/run`

##Environment Variables
The service is built to read configuration out of the environment variables. An example of what variables are configurable can be found [here](https://github.com/evanfarrar/uaa-sso-debug/blob/master/bin/env/example).

There can be multiple environments for the service. The environment can be set by declaring the `ENVIRONMENT` variable, which has a default value of `development`. When running locally, the service will read the enviroment variables out of the matching `bin/env/$ENVIRONMENT` file.

##Running Tests
###Unit Suite
The unit test suite can be run with the `bin/unit` command. This command will target the `test` environment unless another `ENVIRONMENT` has been declared. The unit tests require a database, and will truncate all tables therein. It is recommended that the databases for the `development` and `test` environments are separate.

The following will run the tests for the `./web/middleware` package:

    $ bin/unit ./web/middleware/
    No asset changes
    === RUN TestMiddlewareSuite

    Running Suite: Middleware Suite
    ===============================
    Random Seed: 1403025501
    Will run 15 of 15 specs

    •••••••••••••••
    Ran 15 of 15 Specs in 0.002 seconds
    SUCCESS! -- 15 Passed | 0 Failed | 0 Pending | 0 Skipped

    --- PASS: TestMiddlewareSuite (0.00 seconds)
    PASS
    coverage: 98.2% of statements
    ok    github.com/evanfarrar/uaa-sso-debug/web/middleware 0.071s

    UNIT SUITE PASS

The entire suite can be run by not specifying a directory path: `bin/unit`.

###Javascript Suite
The javascript test suite uses [Karma](https://github.com/karma-runner/karma) to run the tests headlessly in Firefox. The tests can be run by executing the `bin/js test` command. Other information about the `bin/js` command can be found [here](#binjs)

    $ bin/js test
    Running "karma:run" (karma) task
    INFO [karma]: Karma v0.12.10 server started at http://localhost:9876/
    INFO [launcher]: Starting browser Firefox
    INFO [Firefox 29.0.0 (Mac OS X 10.9)]: Connected on socket 5qflJ1xyZ4ns4McxzxxK   with id 7949530
    Firefox 29.0.0 (Mac OS X 10.9): Executed 82 of 82 SUCCESS (0.206 secs / 0.187   secs)

    Done, without errors.

###Acceptance Suite
The Acceptance tests are run with [Capybara](https://github.com/jnicklas/capybara). We used [simple_bdd](https://github.com/robb1e/simple_bdd) to write our feature specs. The tests can be run as follows:

    $ bin/acceptance
    ~/workspace/go/src/github.com/evanfarrar/uaa-sso-debug/acceptance ~/  workspace/go/src/github.com/evanfarrar/uaa-sso-debug
    No asset changes
    ................

    Finished in 19.51 seconds
    16 examples, 0 failures

    Randomized with seed 65354

    ~/workspace/go/src/github.com/evanfarrar/uaa-sso-debug

    INTEGRATION SUITE PASS

##Development Commands
###`bin/js`

This command consists of four subcommands:

#####`bin/js test`
Info found [here](#javascript-suite)

#####`bin/js build`
This command runs jshint, uglify and minifies the javascript into a single file. For CSS we use [pivotal-ui](https://github.com/pivotal-cf/pivotal-ui) which resides in the public directory along with a single css file of our own.  We don't use any minification procedure for css.

#####`bin/js format`
This formats all javascript files according to the [js-beautify's](https://github.com/beautify-web/js-beautify) settings.

#####`bin/js watch`
This watches all js/css files for changes and recompiles the assets as necessary. This can be run with the application server as described [here](#binrun)

* * *

###`bin/run`
This command boots the app server. Each time the app is booted it checks to see if assets need to be recompiled and runs new migrations.

####options
`bin/run` only has one option, watch. This will recompile assets as they change on the filesystem so restarting the server is not necessary.

    $bin/run watch

* * *

###`bin/format`
This runs the [goimports](https://github.com/bradfitz/goimports) on all of the source.  We have changed the defaults to not use tabs.  Settings for `goimports` are in this script.

##Tools
###Migrations
We use [goose](https://bitbucket.org/liamstask/goose) for migrations.  Migrations are stored in the `./models/migrations` directory.  Goose migrations are ordered. To add a new migration, you need to create a new file according to the naming scheme.  If the last migration was `7_added_stuff.sql` you would need to create `8_what_your_migration_does.sql`.  The contents of the files are plain sql and look like the following:

    -- +goose Up
    ALTER TABLE `scheduled_rules` ADD enabled TINYINT(1) NOT NULL DEFAULT 1;

    -- +goose Down
    ALTER TABLE `scheduled_rules` DROP COLUMN enabled;

###gorp
[gorp](https://github.com/go-gorp/gorp) is our database modeling layer.  Our tables are mapped to structs in the models directory.  Each model has an accompanying repo that deals with talking to the database.

###godep
Go dependencies are managed by [godep](https://github.com/tools/godep) Given that this tools docs are currently unsatifactory our use case is as follows:

##### Adding a new dependency
    $ go get github.com/go-gorp/gorp

    # use the package in your code, godep will not pick up packages that are not imported

    # ./... is important to make godep parse your entire project tree
    $ godep save ./...

###Asset Pipeline
We use [grunt](http://gruntjs.com) installed via [npm](https://www.npmjs.org) to handle all js/css related tasks.  The Gruntfile is located under the assets directory. The majority of asset tasks can be run through the [`bin/js` command](#binjs).

##Implementation Details
The application is a REST-ful JSON API with a Backbone front-end. The backend also includes several background workers performing miscellaneous tasks.

###Server
The server code can be found within the `./web` package. The main entry point is through the `router.go` file where URLs are mapped onto HTTP handlers. All of the handlers conform to the http.Handler interface as specified in the golang `http` [documentation](http://golang.org/pkg/net/http/).

#####Stack
The handlers are formed of several pieces of middleware (logging, permissions, etc.) and a single endpoint handler. This grouping is called the Stack.

###Workers
The worker code can be found in the `./workers` package. The workers are responsible for several important pieces of service operation that happen outside of the HTTP request/response cycle. These tasks include gathering metrics for application instance performance, scaling application instances based on these readings, setting configuration values at scheduled times, and cleaning up old data from the database.

#####Exchange
In several cases, the workers need a method to communicate with one another in an unstructured way. There is a very simple pub/sub implementation within the `./exchange` package that handles this communication. Subscribers simply register an interest in a topic that they will then receive messages for. Publishers can send messages for these topics. The entire exchange is built within the golang runtime. There is no persistence, and publishing messages onto the exchange is non-blocking.

##UAA Setup
The autoscaling service needs a UAA client created in advance to function correctly.  This client must have the `cloud_controller.admin` authority, because it needs to be able to scale apps up and down without the user's access token.

    # setup UAAC, the command line for UAA
    gem install cf-uaac
    uaac target uaa.<your_domain>

    # get a token that is capable of creating your new UAA client
    uaac token client get admin --scope "clients.read,clients.write"
    # the admin client secret (client! not admin scim user) is in the CF deployment manifest under uaa properties

    uaac client add autoscaling_service --scope "openid,cloud_controller.permissions,cloud_controller.read,cloud_controller.write" --authorized_grant_types "client_credentials,authorization_code" --authorities "cloud_controller.write,cloud_controller.read,cloud_controller.admin notifications.write critical_notifications.write emails.write" --access_token_validity 3600 --autoapprove true
    # use whatever secret you want

You now have a UAA client (`UAA_CLIENT_ID` environment variable) called `autoscaling_service` with whatever secret you just chose above.  You need to configure the autoscaling app to know the client secret using the `UAA_CLIENT_SECRET` environment variable

##Notifications Integration
The autoscaling service has an integration with the notifications service. This integration is optional. The service can be configured to send notifications by setting the `NOTIFICATIONS_HOST` environment variable. When the variable is empty, or unset, the service will not send notifications. In addition to setting the `NOTIFICATIONS_HOST` variable, the autoscaling service UAA client will need to have the correct scopes to send a notification. The documentation for configuring a notification sending client can be found [here](https://github.com/cloudfoundry-incubator/notifications#send-notifications).
