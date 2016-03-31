package base

import "flag"

var (
	devmodule = flag.Bool("devmodule", false, "开发模式")
)

func IsDevModule() bool {
	return *devmodule
}

func WarpTags(tags []string) []string {
	if tags == nil {
		tags = []string{}
	}
	if IsDevModule() {
		tags = append([]string{"dev"}, tags...)
	} else {
		tags = append([]string{"pro"}, tags...)
	}
	return tags
}
