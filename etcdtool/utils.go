package etcdtool

import (
	"fmt"
)

func buildServiceKeyPrefix(serviceName string, serviceTag string) string {
	return fmt.Sprintf("/ms/registers/%s/%s/", serviceName, serviceTag)
}
