package log

import (
	"context"
	"os"
	"strings"

	"api/pkg/version"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type correlationIDType int

const (
	requestIDKey correlationIDType = iota
	sessionIDKey
)

// Default logger of the system.
var logger *zap.Logger

// code environment
var environment string

var logLevelSeverity = map[string]zapcore.Level{
	"DEBUG":     zapcore.DebugLevel,
	"INFO":      zapcore.InfoLevel,
	"WARNING":   zapcore.WarnLevel,
	"ERROR":     zapcore.ErrorLevel,
	"CRITICAL":  zapcore.DPanicLevel,
	"ALERT":     zapcore.PanicLevel,
	"EMERGENCY": zapcore.FatalLevel,
}

func init() {

	buildVersion := version.Version
	buildHash := version.BuildHash
	buildDate := version.BuildDate

	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))

	if logLevel == "" {
		logLevel = "INFO"
	}

	environment = os.Getenv("ENVIRONMENT")

	if environment == "" {
		environment = "development"
	}

	config := zap.NewProductionEncoderConfig()
	//config.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	//	nanos := t.UnixNano()
	//	millis := nanos / int64(time.Millisecond)
	//	enc.AppendInt64(millis)
	//}
	encoder := zapcore.NewJSONEncoder(config)

	core := zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zap.NewAtomicLevelAt(logLevelSeverity[logLevel]))
	defaultLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	defer defaultLogger.Sync()

	logger = defaultLogger.With(zap.String("v", buildVersion), zap.String("bh", buildHash), zap.String("bd", buildDate), zap.String("env", environment))
}

// LoggerWithLambdaRqID returns a logger with lambda context
func LoggerWithLambdaRqID(ctx context.Context) *zap.Logger {

	lc, _ := lambdacontext.FromContext(ctx)

	rqCtx := WithRqID(ctx, lc.AwsRequestID)

	return Logger(rqCtx)
}

// WithRqID returns a context which knows its request ID
func WithRqID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// Logger returns a zap logger with as much context as possible
func Logger(ctx context.Context) *zap.Logger {

	newLogger := logger

	if ctx == nil {
		return newLogger
	}

	if ctxRqID, ok := ctx.Value(requestIDKey).(string); ok {
		newLogger = newLogger.With(zap.String("rqID", ctxRqID))
	}

	return newLogger
}
