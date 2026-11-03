package constants

import (
	dawn "github.com/dawnim/dawn-server/constants"
	"time"
)

// TempGreeting demonstrates a business cache visible in the cache manager.
var TempGreeting = dawn.Definition{
	Name: "TempGreeting", Pattern: "temp:greeting:%s", Type: "string",
	Description: "实时示例问候缓存", TTL: 5 * time.Minute, Readable: true, Clearable: true,
}
