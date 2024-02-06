package service

type GetSubService struct {
	subService SubService
}

func (s *GetSubService) GetLatestSubs() string {
	return ""
}
