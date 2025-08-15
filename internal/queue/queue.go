package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	EnqueueActivationEmail(userEmail, userName, token string) error
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
	payload := map[string]string{
		"userEmail": userEmail,
		"userName":  userName,
		"token":     token,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Error marshaling activation email payload: %v", err)
		return err
	}

	_, err = qs.asynqClient.Enqueue(asynq.NewTask("email:send_activation", payloadBytes), asynq.Queue("emails"))
	if err != nil {
		log.Errorf("Error queuing activation email task: %v", err)
		return err
	}

	return nil
}