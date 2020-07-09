package plugin

import "context"

func EnablePlugins(ctx context.Context, enables ...func(ctx context.Context)) {
	for _, enableFunc := range enables {
		enableFunc(ctx)
	}
}
