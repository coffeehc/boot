package grpcrecovery

import (
	"context"
	stderrors "errors"
	"io"

	"github.com/coffeehc/base/errors"
	"github.com/coffeehc/base/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

func init() {
	grpclog.SetLoggerV2(NewZapLogger())
	grpc.EnableTracing = false
}

var errCode = codes.Code(18)

// convertRPCError 保留标准 gRPC 错误，其他错误使用 Boot 已有格式传输；恢复路径不得再次 panic。
func convertRPCError(err interface{}, recover bool, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	// 保留标准状态码和取消语义；18 只用于已有的 base/errors 传输格式。
	if v, ok := err.(error); ok && !recover {
		if stderrors.Is(v, context.Canceled) || stderrors.Is(v, context.DeadlineExceeded) {
			return status.FromContextError(v).Err()
		}
		if _, ok := status.FromError(v); ok {
			return v
		}
	}
	var errs errors.Error
	switch v := err.(type) {
	case errors.Error:
		if errors.IsSystemError(v) {
			if !DisableGrpcLog {
				log.Error(v.Error(), v.GetFields()...)
			}
		}
		errs = v
	case string:
		if recover && !DisableGrpcLog {
			log.Error("RPC 调用发生 panic", append(fields, zap.String("error", v))...)
		}
		errs = errors.SystemError(v)
	case error:
		errs = errors.SystemError(v.Error())
		if recover && !DisableGrpcLog {
			log.Error("RPC 调用发生 panic", append(fields, zap.Error(v))...)
		}
	default:
		if !DisableGrpcLog {
			log.Error("RPC 调用发生 panic", append(fields, zap.Any("err", v))...)
		}
		errs = errors.SystemError("未知异常")
	}
	return status.Error(errCode, errs.FormatRPCError())
}

// parseRPCError 解码 Boot 错误，保留标准状态码和流结束语义；日志开关不影响返回值。
func parseRPCError(err interface{}, recover bool, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case errors.Error:
		return v
	case string:
		if !DisableGrpcLog {
			if recover {
				log.Error("RPC 客户端发生 panic", append(fields, zap.String("error", v))...)
			} else {
				log.Warn("rpc错误", append(fields, zap.String("err", v))...)
			}
		}
		return errors.SystemError(v)
	case error:
		if recover {
			if !DisableGrpcLog {
				log.Error("RPC 客户端发生 panic", append(fields, zap.Error(v))...)
			}
			return errors.SystemError(v.Error())
		}
		if stderrors.Is(v, io.EOF) || stderrors.Is(v, context.Canceled) || stderrors.Is(v, context.DeadlineExceeded) {
			return v
		}
		s, ok := status.FromError(v)
		if !ok {
			if !DisableGrpcLog {
				log.Error("RPC 客户端调用失败", append(fields, zap.Error(v))...)
			}
			return errors.WrappedSystemError(v)
		}
		if s.Code() == errCode {
			return errors.ParseError(s.Message())
		}
		return v
	}
	if !DisableGrpcLog {
		log.Error("RPC 客户端发生未知异常", append(fields, zap.Any("err", err))...)
	}
	return errors.SystemError("未知异常")
}
