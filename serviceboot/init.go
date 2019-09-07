package serviceboot

import (
	"git.xiagaogao.com/coffee/boot/bootconfig"
	"github.com/json-iterator/go/extra"
)

func init() {
	extra.RegisterFuzzyDecoders()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
	bootconfig.InitConfig()
}
