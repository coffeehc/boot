package etcdtool

import (
	"fmt"
)

func buildServiceKeyPrefix(serviceName string) string {
	return fmt.Sprintf("internal_ms.%s.", serviceName)
}
