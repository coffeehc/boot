package grpcrecovery

import (
	"strings"

	"git.xiagaogao.com/coffee/base/errors"
	"git.xiagaogao.com/coffee/base/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errCode = codes.Code(18)

func convertRPCError(err interface{}, recover bool, fields ...zap.Field) error {
	if err == nil {
		return nil
	}
	var errs errors.Error = nil
	switch v := err.(type) {
	case errors.Error:
		if errors.IsSystemError(v) {
			log.DPanic(v.Error(), v.GetFields()...)
		} else {
			if !strings.HasPrefix(v.Error(), "context") {
				log.Error(v.Error(), v.GetFields()...)
			}
		}
		errs = v
	case string:
		if recover {
			log.DPanic("不可处理的异常", append(fields, zap.String("error", v))...)
		}
		errs = errors.SystemError(v)
	case error:
		if recover {
			log.DPanic("不可处理的异常", append(fields, zap.Error(v))...)
		} else {
			errs = errors.SystemError(v.Error())
		}
	default:
		log.DPanic("不可处理的异常", append(fields, zap.Any("err", v))...)
		errs = errors.SystemError("未知异常")
	}
	return status.Errorf(errCode, errs.FormatRPCError())
}

func parseRPCError(err interface{}, recover bool, fields ...zap.Field) errors.Error {
	if err == nil {
		return nil
	}
	switch v := err.(type) {
	case errors.Error:
		return v
	case string:
		if recover {
			log.DPanic("不可处理的异常", append(fields, zap.String("error", v))...)
		}
		return errors.SystemError(v)
	case error:
		s, ok := status.FromError(v)
		if !ok {
			log.DPanic("无法识别的RPC异常", append(fields, zap.Error(v))...)
			return errors.SystemError("无法识别的RPC异常")
		}
		if s.Code() == errCode {
			return errors.ParseError(s.Message())
		}
		if recover {
			log.DPanic("不可处理的异常", append(fields, zap.Error(v))...)
		} else {
			// log.Warn("远程服务暂时不可用", append(fields, zap.Error(v))...)
		}
		return errors.SystemError("远程服务暂时不可用,请重试")
	}
	log.DPanic("未知异常", append(fields, zap.Any("err", err))...)
	return errors.SystemError("未知异常")
}
