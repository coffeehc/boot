package grpcrecovery

import (
	"testing"

	"github.com/coffeehc/base/errors"
	"google.golang.org/grpc/status"
)

func TestConvertRPCErrorPreservesLiteralPercent(t *testing.T) {
	cause := errors.MessageError("进度 50%，占位符 %s 应原样保留")
	converted := convertRPCError(cause, false)
	if got := status.Convert(converted); got.Code() != errCode || got.Message() != cause.FormatRPCError() {
		t.Fatalf("converted status=%v, want literal message %q", got, cause.FormatRPCError())
	}
}
