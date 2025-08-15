package queue

import (
	"encoding/json"
	"ypeskov/budget-go/internal/constants"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

type ActivationEmailPayload struct {
	UserEmail string `json:"userEmail"`
	UserName  string `json:"userName"`
	Token     string `json:"token"`
}

type Service interface {
	EnqueueActivationEmail(userEmail, userName, token string) error
	EnqueueDBBackup() error
	EnqueueExchangeRatesUpdate() error
}

type ServiceInstance struct {
	asynqClient *asynq.Client
}

func NewService(asynqClient *asynq.Client) Service {
	return &ServiceInstance{
		asynqClient: asynqClient,
	}
}

func (qs *ServiceInstance) EnqueueActivationEmail(userEmail, userName, token string) error {
	payload := ActivationEmailPayload{
		UserEmail: userEmail,
		UserName:  userName,
		Token:     token,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling activation email payload: %v", err)
		return err
	}

	_, err = qs.asynqClient.Enqueue(asynq.NewTask(constants.TaskSendActivationEmail, payloadBytes), asynq.Queue("emails"))
	if err != nil {
		log.Errorf("Error queuing activation email task: %v", err)
		return err
	}

	return nil
}

func (qs *ServiceInstance) EnqueueDBBackup() error {
	_, err := qs.asynqClient.Enqueue(asynq.NewTask(constants.TaskDBBackupDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("Error queuing DB backup task: %v", err)
		return err
	}
	return nil
}

func (qs *ServiceInstance) EnqueueExchangeRatesUpdate() error {
	_, err := qs.asynqClient.Enqueue(asynq.NewTask(constants.TaskExchangeRatesDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("Error queuing exchange rates update task: %v", err)
		return err
	}
	return nil
}