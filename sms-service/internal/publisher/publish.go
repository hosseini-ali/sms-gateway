package publisher

import (
	"context"
	"notif/internal/models"
)


type Publisher interface {
	Publish(ctx context.Context, event models.SMSLog) error
}

