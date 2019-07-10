package serviceboot

import "github.com/json-iterator/go/extra"

func init() {
	extra.RegisterFuzzyDecoders()
	extra.SetNamingStrategy(extra.LowerCaseWithUnderscores)
}
