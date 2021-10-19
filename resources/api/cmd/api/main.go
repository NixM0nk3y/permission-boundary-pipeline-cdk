package main

import (
	"context"

	"api/internal/api"
	"api/internal/config"

	"api/pkg/log"
	"api/pkg/log/chilogger"
	"api/pkg/version"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-xray-sdk-go/xray"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

// our router
var chiLambda *chiadapter.ChiLambda

func init() {

	logger := log.Logger(context.TODO())

	// stdout and stderr are sent to AWS CloudWatch Logs
	logger.Warn("lambda cold start")

	r := chi.NewRouter()

	// various middlewares
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(chilogger.Logger())
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/version", version.Handler)

	r.Get("/hello", api.HelloWorldHandler)

	chiLambda = chiadapter.New(r)
}

// Handler is
func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	logger := log.LoggerWithLambdaRqID(ctx)

	xray.SetLogger(&log.XrayLogger{})

	xray.Configure(xray.Config{
		LogLevel:       "warn",
		ServiceVersion: version.Version,
	})

	logger.Info("application handler")

	logger.Debug("recieved event", zap.Reflect("req", req))

	vctx := config.ReadEnvConfig(ctx, "APPLICATION")

	return chiLambda.ProxyWithContext(vctx, req)
}

func main() {
	lambda.Start(Handler)
}
