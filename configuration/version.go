package configuration

import "fmt"

var (
	// BuildVersion = ""
	GitRev    = ""
	GitTag    = ""
	BuildTime = ""
	Version   = ""
)

func PrintVersionInfo() {
	fmt.Printf("版本信息:%s\n", BuildTime)
	fmt.Printf("\tBuildTime:%s\n", BuildTime)
	fmt.Printf("\tGitRev:%s\n", GitRev)
	fmt.Printf("\tGitTag:%s\n", GitTag)
	fmt.Printf("\tVersion:%s\n", Version)
}
