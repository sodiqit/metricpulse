package services

type IAppService interface {
	SayHello() string
}

type AppService struct {
}

func (service AppService) SayHello() string {
	return "Hello world!"
}

func NewAppService() AppService {
	return AppService{}
}
