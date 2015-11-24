package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	ScaleDownNotification = NotificationRegistration{
		KindID:      "11c241c9-bf40-4d7a-93c1-ecf2db115fa4",
		Description: "Scaling Down",
		Subject:     "Scaling Down",
	}

	ScaleUpNotification = NotificationRegistration{
		KindID:      "effa96de-2349-423a-b5e4-b1e84712a714",
		Description: "Scaling Up",
		Subject:     "Scaling Up",
	}

	ManualScalingNotification = NotificationRegistration{
		KindID:      "4605d2d4-542c-4f54-b589-cdaa0078c3e1",
		Description: "Manual Scaling Detected",
		Subject:     "Manual Scaling Detected",
	}

	MaxInstanceCountNotification = NotificationRegistration{
		KindID:      "4fb85f46-c52b-4dc3-8bcc-8d8394f39b77",
		Description: "Maximum Instance Limit Reached",
		Subject:     "Maximum Instance Limit Reached",
	}

	QuotaLimitNotification = NotificationRegistration{
		KindID:      "78740d82-080d-4162-bf4c-9607ae635b53",
		Description: "Quota Limit Reached",
		Subject:     "Quota Limit Reached",
	}
)

type NotificationRegistration struct {
	KindID      string
	Description string
	Subject     string
}

type NotificationsRequestError struct {
	statusCode   int
	errorMessage string
}

func NewNotificationsRequestError(statusCode int, errorMessage string) NotificationsRequestError {
	return NotificationsRequestError{
		statusCode:   statusCode,
		errorMessage: errorMessage,
	}
}

func (err NotificationsRequestError) Error() string {
	return fmt.Sprintf("Notifications request failed: StatusCode: %d, Error: %s", err.statusCode, err.errorMessage)
}

type NotificationsClientInterface interface {
	Notify(SpaceNotification) error
}

type SpaceNotification struct {
	SpaceGUID string
	KindID    string
	Subject   string
	Text      string
	HTML      string
}

type NotificationsClient struct {
	host      string
	uaaClient UAAInterface
	logger    *log.Logger
}

type Registration struct {
	SourceDescription string `json:"source_description"`
	Kinds             []Kind `json:"kinds"`
}

type Kind struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func NewNotificationsClient(url string, uaaClient UAAInterface, logger *log.Logger) NotificationsClient {
	return NotificationsClient{
		host:      url,
		uaaClient: uaaClient,
		logger:    logger,
	}
}

func (client NotificationsClient) Notify(spaceNotification SpaceNotification) error {
	if client.host == "" {
		client.logger.Println("Notifications Client not configured")
		return nil
	}

	params := map[string]string{
		"kind_id": spaceNotification.KindID,
		"text":    spaceNotification.Text,
		"html":    spaceNotification.HTML,
		"subject": spaceNotification.Subject,
	}

	body, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	request, err := http.NewRequest("POST", client.host+"/spaces/"+spaceNotification.SpaceGUID, bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}

	token, err := client.uaaClient.GetClientToken()
	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+token.Access)

	response, err := GetClient().Do(request)
	if err != nil {
		panic(err)
	}

	if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return NewNotificationsRequestError(response.StatusCode, string(body))
	}
	return nil
}

func (client NotificationsClient) Register() error {
	if client.host == "" {
		client.logger.Println("Notifications Client not configured")
		return nil
	}

	registration := Registration{
		SourceDescription: "Cloud Foundry Autoscaling Service",
		Kinds: []Kind{
			{
				ID:          ScaleDownNotification.KindID,
				Description: ScaleDownNotification.Description,
			},
			{
				ID:          ManualScalingNotification.KindID,
				Description: ManualScalingNotification.Description,
			},
			{
				ID:          MaxInstanceCountNotification.KindID,
				Description: MaxInstanceCountNotification.Description,
			},
			{
				ID:          ScaleUpNotification.KindID,
				Description: ScaleUpNotification.Description,
			},
			{
				ID:          QuotaLimitNotification.KindID,
				Description: QuotaLimitNotification.Description,
			},
		},
	}

	body, err := json.Marshal(registration)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", client.host+"/registration", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	token, err := client.uaaClient.GetClientToken()
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+token.Access)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return NewNotificationsRequestError(response.StatusCode, string(body))
	}

	return nil
}
