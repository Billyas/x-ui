package job

import "x-ui/web/service"

type GetSubsJob struct {
	getSubService service.GetSubService
}

func NewGetSubsJob() *GetSubsJob {
	return new(GetSubsJob)
}

func (j *GetSubsJob) Run() {
	_, err := j.getSubService.GetLatestUrlSub()
	if err != nil {
		return
	}
}
