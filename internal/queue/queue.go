package queue

import (
	"encoding/json"
	"sync"
	"ypeskov/budget-go/internal/constants"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

type ActivationEmailPayload struct {
	UserEmail string `json:"userEmail"`
	UserName  string `json:"userName"`
	Token     string `json:"token"`
}

type QueueService interface {
	EnqueueActivationEmail(userEmail, userName, token string) error
	EnqueueDBBackup() error
	EnqueueExchangeRatesUpdate() error
}

type QueueServiceInstance struct {
	asynqClient *asynq.Client
}

var (
	queueServiceInstance *QueueServiceInstance
	queueOnce            sync.Once
)

func NewQueueService(asynqClient *asynq.Client) QueueService {
	queueOnce.Do(func() {
		log.Debug("Creating Queue service instance")
		queueServiceInstance = &QueueServiceInstance{
			asynqClient: asynqClient,
		}
	})

	return queueServiceInstance
}

func (qs *QueueServiceInstance) EnqueueActivationEmail(userEmail, userName, token string) error {
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

func (qs *QueueServiceInstance) EnqueueDBBackup() error {
	_, err := qs.asynqClient.Enqueue(asynq.NewTask(constants.TaskDBBackupDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("Error queuing DB backup task: %v", err)
		return err
	}
	return nil
}

func (qs *QueueServiceInstance) EnqueueExchangeRatesUpdate() error {
	_, err := qs.asynqClient.Enqueue(asynq.NewTask(constants.TaskExchangeRatesDaily, nil), asynq.Queue("default"))
	if err != nil {
		log.Errorf("Error queuing exchange rates update task: %v", err)
		return err
	}
	return nil
}