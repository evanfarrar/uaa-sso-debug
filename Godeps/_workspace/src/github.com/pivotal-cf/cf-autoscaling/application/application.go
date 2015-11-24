package application

import (
	"time"

	"github.com/evanfarrar/uaa-sso-debug/config"
	"github.com/evanfarrar/uaa-sso-debug/exchange"
	"github.com/evanfarrar/uaa-sso-debug/log"
	"github.com/evanfarrar/uaa-sso-debug/models"
	"github.com/evanfarrar/uaa-sso-debug/services"
	"github.com/evanfarrar/uaa-sso-debug/util"
	"github.com/evanfarrar/uaa-sso-debug/web"
	"github.com/evanfarrar/uaa-sso-debug/workers"
	"github.com/evanfarrar/uaa-sso-debug/workers/metrics"
	"github.com/evanfarrar/uaa-sso-debug/workers/scaling"
	"github.com/evanfarrar/uaa-sso-golang/uaa"
	"github.com/ryanmoran/viron"
)

type Application struct{}

func NewApplication() Application {
	return Application{}
}

func (app Application) PrintEnvironment() {
	env := config.NewEnvironment()
	viron.Print(env, log.Logger)
}

func (app Application) StartWorkers() {
	env := config.NewEnvironment()

	if env.VCAPApplication.InstanceIndex == 0 {
		messageExchange := exchange.New()
		scalingDecisionsRepo := models.NewScalingDecisionsRepo()
		serviceBindingsRepo := models.NewServiceBindingsRepo()
		readingsRepo := models.NewReadingsRepo()
		serviceInstancesRepo := models.NewServiceInstancesRepo()
		plansRepo := models.NewPlansRepo()
		keyValueRepo := models.NewKeyValueRepo()
		scheduledRulesRepo := models.NewScheduledRulesRepo()
		cloudControllerClient := services.NewCloudControllerClient()
		notificationsClient := services.NewNotificationsClient(env.NotificationsHost, app.UAAClient(env), log.Logger)
		err := notificationsClient.Register()
		if err != nil {
			log.Printf("Notifications registration failed: %s", err.Error())
		}

		metricsPoller := metrics.NewPoller(serviceBindingsRepo, time.Duration(env.MetricsPollInterval)*time.Second)
		metricsDirector := metrics.NewDirector(serviceBindingsRepo, serviceInstancesRepo, readingsRepo, messageExchange, metricsPoller)
		scalingDirector := scaling.NewDirector(messageExchange, serviceBindingsRepo, scalingDecisionsRepo)

		scalingReportFactory := scaling.NewScalingReportFactory(scalingDecisionsRepo, serviceBindingsRepo, serviceInstancesRepo, plansRepo, readingsRepo)
		notifier := scaling.NewNotifier(scalingDecisionsRepo, scalingReportFactory, cloudControllerClient, notificationsClient)
		schedulingDirector := workers.NewSchedulingDirector(keyValueRepo, scheduledRulesRepo, serviceBindingsRepo, scalingDecisionsRepo, util.NewClock(), util.NewTimer(1*time.Minute))
		garbageCollector := workers.NewGarbageCollector(5*time.Minute, readingsRepo, scalingDecisionsRepo, scheduledRulesRepo, keyValueRepo)

		metricsDirector.Run()
		scalingDirector.Run()
		notifier.Run()
		schedulingDirector.Run()
		garbageCollector.Run()
	}
}

func (app Application) UAAClient(env config.Environment) uaa.UAA {
	auth := uaa.NewUAA(env.LoginHost, env.UAAHost, env.UAAClientID, env.UAAClientSecret, "")
	auth.RedirectURL = env.Domain + "/sessions/create"
	auth.Scope = "openid,cloud_controller.permissions,cloud_controller.read,cloud_controller.write"
	auth.AccessType = "offline"
	auth.VerifySSL = env.VerifySSL

	return auth
}

func (app Application) StartServer() {
	server := web.NewServer()
	server.Start()
}

func (app Application) Crash() {
	err := recover()
	if err != nil {
		time.Sleep(10 * time.Second)
		panic(err)
	}
}
