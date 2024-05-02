package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	middleware "github.com/oapi-codegen/gin-middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robertjshirts/gobuster/api"
	"github.com/robertjshirts/gobuster/dal"
	"github.com/robertjshirts/gobuster/services"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"code"})
	avgResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "api_http_response_time_seconds",
		Help: "Average response time of HTTP requests, sorted by endpoint",
	}, []string{"endpoint"})
	serverErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "api_server_errors_total",
		Help: "Total number of server errors",
	}, []string{"error"})
)

func RequestCounterMiddleware(c *gin.Context) {
	c.Next()
	httpRequestsTotal.WithLabelValues(fmt.Sprint(c.Writer.Status())).Inc()
}

func ResponseTimeMiddleware(c *gin.Context) {
	start := time.Now()
	c.Next()
	elapsed := time.Since(start)
	avgResponseTime.WithLabelValues(c.FullPath()).Observe(elapsed.Seconds())
}

func ErrorLoggingMiddleware(c *gin.Context) {
	c.Next()
	for _, e := range c.Errors {
		serverErrorsTotal.WithLabelValues(e.Error()).Inc()
		log.Println(e)
	}
}

func main() {
	router := gin.Default()
	router.Use(RequestCounterMiddleware)
	router.Use(ResponseTimeMiddleware)
	router.Use(ErrorLoggingMiddleware)

	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(avgResponseTime)
	prometheus.MustRegister(serverErrorsTotal)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	dbConfig := ReadDatabaseConfig("config/database.config")
	kafkaConfig := ReadSaramaConfig("config/kafka.config")

	db, dbErr := dal.Init(dbConfig["user"], dbConfig["password"], dbConfig["protocol"], dbConfig["host"], dbConfig["port"], dbConfig["database"])
	if dbErr != nil {
		log.Fatal("There was an error connecting to the database: \n", dbErr)
	}
	defer db.Close()

	service, sErr := services.Init(db, strings.Split(kafkaConfig["brokers"], ","), kafkaConfig["offerTopic"], kafkaConfig["userTopic"])
	if sErr != nil {
		log.Fatal("There was an error initializing the services: \n", sErr)
	}
	defer service.Close()

	router.Use(service.Middleware)

	si := api.Init(service)

	swagger, sErr := api.GetSwagger()
	if sErr != nil {
		log.Fatal(sErr)
	}

	router.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(router, si)
	router.Run()
}
