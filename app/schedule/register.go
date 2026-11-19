package schedule

import (
	"context"
	"github.com/dawnim/dawn-server/constants"
	"github.com/dawnim/dawn-simple/app/web"
	"github.com/prismgo/framework/queue"
	"github.com/prismgo/framework/timer"
)

// Register 在这里注册项目定时任务。
func Register(s *timer.Schedule) {
	s.Call(func(ctx context.Context) error {
		_, err := queue.Dispatch(ctx, &web.AutoMomentJob{}, queue.OnQueue(constants.QueueDefault), queue.Tries(1))
		return err
	}).EverySixHours().Name("moments.auto")
}
