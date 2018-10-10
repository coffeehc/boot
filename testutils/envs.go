package testutils

import "os"

func SetENVRunModel(model string) {
	os.Setenv("ENV_RUN_MODEL", model)
}

func SetLoggerLevel(level string) {
	os.Setenv("ENV_LOGGER_LEVEL", level)
}
