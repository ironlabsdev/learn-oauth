package ctx

import (
	"context"

	"oauth/utils/env"
)

const keyRequestID key = "requestID"

const envConfigKey key = "env_config"

const userIdKey key = "user_id"

type key string

func UserID(ctx context.Context) int32 {
	userID, _ := ctx.Value(userIdKey).(int32)

	return userID
}

func SetUserID(ctx context.Context, userID int32) context.Context {
	return context.WithValue(ctx, userIdKey, userID)
}

func RequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(keyRequestID).(string)

	return requestID
}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, keyRequestID, requestID)
}

func EnvConfigID(ctx context.Context) *env.Conf {
	config, _ := ctx.Value(envConfigKey).(*env.Conf)
	return config
}

func SetEnvConfigID(ctx context.Context, conf *env.Conf) context.Context {
	return context.WithValue(ctx, envConfigKey, conf)
}
