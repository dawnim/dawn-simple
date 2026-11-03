package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	sdk "github.com/dawnim/dawn-server/app/admin"
	"github.com/dawnim/dawn-server/realtime"
	"github.com/dawnim/dawn-server/support/i18n"
	appconst "github.com/dawnim/dawn-simple/app/constants"
	"github.com/redis/go-redis/v9"
)

// RegisterRealtime shares authentication, connections, Redis and application lifecycle.
func RegisterRealtime(a *sdk.App, hub *realtime.Hub) error {
	return hub.Handle("app.greeting", func(ctx context.Context, p realtime.Principal, _ json.RawMessage) (any, error) {
		if p.Realm != "admin" {
			return nil, i18n.Errorf("没有操作权限")
		}
		actor, err := a.ActorForSession(p.SessionID)
		if err != nil {
			return nil, err
		}
		if !actor.Can("app.realtime_example.view") {
			return nil, i18n.Errorf("没有操作权限")
		}
		key := a.Key(appconst.TempGreeting, p.Subject)
		greeting, err := a.Redis.Get(ctx, key).Result()
		if errors.Is(err, redis.Nil) {
			greeting = fmt.Sprintf("Hello, %s", actor.User.Username)
			err = a.Redis.Set(ctx, key, greeting, appconst.TempGreeting.TTL).Err()
		}
		return map[string]string{"greeting": greeting}, err
	})
}
