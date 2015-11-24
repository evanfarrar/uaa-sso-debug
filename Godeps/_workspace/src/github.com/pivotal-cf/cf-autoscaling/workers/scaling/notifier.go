package scaling

import (
	"time"

	"github.com/evanfarrar/uaa-sso-debug/log"

	"github.com/evanfarrar/uaa-sso-debug/services"
)

var PollingWaitDuration = 10 * time.Second

type Notifier struct {
	scalingDecisionsRepo models.ScalingDecisionsInterface
	factory              ScalingReportFactoryInterface
	cc                   services.CloudControllerInterface
	notificationsClient  services.NotificationsClientInterface
	halt                 chan bool
}

func NewNotifier(scalingDecisionsRepo models.ScalingDecisionsInterface, factory ScalingReportFactoryInterface,
	cc services.CloudControllerInterface, notificationsClient services.NotificationsClientInterface) Notifier {
	return Notifier{
		scalingDecisionsRepo: scalingDecisionsRepo,
		factory:              factory,
		cc:                   cc,
		notificationsClient:  notificationsClient,
		halt:                 make(chan bool),
	}
}

func (worker Notifier) Run() {
	go func() {
		for {
			select {
			case <-worker.halt:
				return
			default:
				worker.Execute()
			}
		}
	}()
}

func (worker Notifier) Halt() {
	worker.halt <- true
}

func (worker Notifier) Execute() {
	decision, err := worker.scalingDecisionsRepo.FindFirstNonNotified()
	if err != nil {
		if err == models.ErrRecordNotFound {
			<-time.After(PollingWaitDuration)
			return
		}
		panic(err)
	}

	report := worker.factory.NewScalingReport(decision)
	if !decision.IsIgnored() {
		notification := report.BuildNotification()
		log.Printf("[Notifier] Notification -> %+v\n", notification)
		if report.ShouldSend() {
			log.Println("[Notifier] Sending notification")
			summary, err := worker.cc.AppSummary(report.Binding().AppGuid)
			if err == nil {
				log.Printf("[Notifier] Retrieved app summary -> %+v\n", summary)
				notification.SpaceGUID = summary.SpaceGUID
				err = worker.notificationsClient.Notify(notification)
				if err != nil {
					log.Printf("[Notifier] Notifications Error encountered: %+v\n", err)
				}
			} else {
				log.Printf("[Notifier] CloudController Error encountered: %+v\n", err)
				if err == services.CCErrors.Failure {
					return
				}
			}
		}
	}

	decision.Notified = true
	_, err = worker.scalingDecisionsRepo.Update(decision)
	if err != nil {
		panic(err)
	}
}
