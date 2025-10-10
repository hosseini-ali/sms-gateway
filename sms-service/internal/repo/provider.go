package repo

import "time"


type Provider interface {
	SendSMS(phoneNumber string, message string) error
}


type IrancelService struct {

}

func (i IrancelService) SendSMS(phoneNumber string, message string) error {
	time.Sleep(1 * time.Second)
	return nil
}

func NewSMSProvider() IrancelService {
	return IrancelService{}
}