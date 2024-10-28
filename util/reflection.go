package util

import (
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime"
	"strings"
)

func platformNotSupported(target any) {
	funcName := getFunctionName(target)
	slog.Error(fmt.Sprintf("[%s] Function not supported on platform '%s'", funcName, runtime.GOOS))
	os.Exit(1)
}

func getFunctionName(i interface{}) string {
	parts := strings.Split(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name(), "/")
	return strings.Join(parts[3:], "/")
}
