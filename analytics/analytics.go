package analytics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type EventHandler struct {
	prefix    string
	ddbClient *dynamodb.DynamoDB
}

func NewAnalyticsHandler(prefix string) (*EventHandler, error) {

	var sess *session.Session

	config := aws.Config{Region: aws.String("us-east-1")}

	if _, isNow := os.LookupEnv("NOW"); isNow {
		// Need to perform a workaround - see https://vercel.com/docs/v2/platform/limits#reserved-variables
		config.Credentials = credentials.NewStaticCredentials(os.Getenv("NOW_AWS_ACCESS_KEY_ID"), os.Getenv("NOW_AWS_SECRET_ACCESS_KEY"), "")
	}

	sess = session.Must(session.NewSession(&config))

	ddb := dynamodb.New(sess)

	return &EventHandler{prefix: prefix, ddbClient: ddb}, nil
}

func (a *EventHandler) GetEngine() *gin.Engine {
	router := gin.Default()

	/*
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowCredentials = true
		corsConfig.AllowOrigins = []string{"https://cdpn.io"}

		corsMiddleware := cors.New(corsConfig)

		router.Use(corsMiddleware)
	*/

	apiRoot := router.Group(a.prefix)

	apiRoot.GET("/", a.RootHandler)
	apiRoot.Any("/tp2", a.SnowplowHandler)

	return router
}

func (a *EventHandler) RootHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (a *EventHandler) SnowplowHandler(c *gin.Context) {
	// Sets the User Identification Cookie
	if _, err := c.Cookie("sp"); err == http.ErrNoCookie {
		c.SetCookie("sp", uuid.New().String(), 86400*365*2, "/", "", false, false)
	}

	if c.Request.Method == "OPTIONS" {
		a.PreflightHandler(c)

		return
	}

	rec := &EventsRecord{}

	if err := c.ShouldBindJSON(rec); err == nil {
		log.Infof("Processing: %+v", rec)

		err = a.handleRecords(rec)

		if nil != err {
			log.Warnf("Oops: %s", err)
		}
	} else {
		log.Warnf("Oops: %s", err)
	}

	c.Data(200, "text/plain", []byte("ok"))
}

// handleRecords writes into event table
func (a *EventHandler) handleRecords(eventArray *EventsRecord) error {
	var writeRequests []*dynamodb.WriteRequest

	for _, v := range eventArray.Data {
		writeReq, err := v.AsWriteRequest()

		if nil != err {
			log.Warnf("Oops: %s", err)
			continue
		}

		writeRequests = append(writeRequests, writeReq)
	}

	_, err := a.ddbClient.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]*dynamodb.WriteRequest{
			"events": writeRequests,
		},
	})

	return err
}

func (a *EventHandler) PreflightHandler(c *gin.Context) {
	a.SetHeaders(c)

	c.String(200, "%s", "text/plain")
}

func (a *EventHandler) SetHeaders(c *gin.Context) {
	origin := "https://cdpn.io"

	if originHeaderValue := c.GetHeader("origin"); originHeaderValue != "" {
		origin = originHeaderValue
	}

	headers := map[string]string{
		"Access-Control-Allow-Origin":      origin,
		"Access-Control-Allow-Headers":     "Access-Control-Allow-Origin,Origin,Content-Length,Content-Type",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET,POST",
		"Access-Control-Allow-Max-Age":     "43200",
	}

	for k, v := range headers {
		c.Header(k, v)
	}
}
