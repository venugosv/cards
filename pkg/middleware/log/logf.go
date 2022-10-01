package logf

import (
	"context"
	"fmt"
	"runtime"

	"github.com/anzx/pkg/monitoring/attributes"

	"github.com/anzx/pkg/log"
)

// This file exists as anzx/pkg/log does not have a sprintf function/interface
// The runtime.Caller and codeurl part is to ensure the code.url attribute points to the caller of this file
// And NOT the function itself
// TODO: When we upgrade Golang to a version with "any" support we should change interface{} to any

func Debug(ctx context.Context, msg string, a ...interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Debug(ctx, fmt.Sprintf(msg, a...), log.Any(string(attributes.CodeURL), attributes.CreateCodeURL(f, l)))
}

func Error(ctx context.Context, err error, msg string, a ...interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Error(ctx, err, fmt.Sprintf(msg, a...), log.Any(string(attributes.CodeURL), attributes.CreateCodeURL(f, l)))
}

func Err(ctx context.Context, err error) {
	_, f, l, _ := runtime.Caller(1)
	log.Error(ctx, err, "", log.Any(string(attributes.CodeURL), attributes.CreateCodeURL(f, l)))
}

func Info(ctx context.Context, msg string, a ...interface{}) {
	_, f, l, _ := runtime.Caller(1)
	log.Info(ctx, fmt.Sprintf(msg, a...), log.Any(string(attributes.CodeURL), attributes.CreateCodeURL(f, l)))
}
