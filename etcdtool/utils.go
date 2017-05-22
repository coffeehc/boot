package etcdtool

import (
	"fmt"
)

func buildServiceKeyPrefix(serviceName string) string {
	return fmt.Sprintf("/ms/registers/%s/", serviceName)
}
