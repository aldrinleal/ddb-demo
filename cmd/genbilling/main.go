package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"strconv"
	"strings"
)

func main() {
	cfg := aws.NewConfig().WithRegion("us-east-1")

	sess := session.Must(session.NewSession(cfg))

	ddb := dynamodb.New(sess)

	scanInput := &dynamodb.ScanInput{
		TableName: aws.String("summary"),
		Limit:     aws.Int64(1),
	}

	for {
		scanOutput, err := ddb.Scan(scanInput)

		if nil != err {
			log.Fatalln(err)
		}

		if 0 == len(scanOutput.LastEvaluatedKey) {
			break
		}

		for _, i := range scanOutput.Items {
			domain_md5 := *(i["domain_md5"].S)

			elements_domain := strings.SplitN(domain_md5, "-", 2)

			domain := elements_domain[1]

			datepath := *(i["datepath"].S)

			e := strings.SplitN(datepath, ":", 2)

			date := e[0]
			path := e[1]

			delete(i, "datepath")
			delete(i, "domain_md5")
			delete(i, "ttl")

			for k, v := range i {
				vAsNumber, err := strconv.ParseInt(*v.N, 10, 64)

				record := map[string]interface{}{
					"domain": domain,
					"date":   date,
					"path":   path,
					"time":   k,
					"hits":   vAsNumber,
				}

				byteArr, err := json.Marshal(&record)

				if nil != err {
					panic(err)
				}

				fmt.Println(string(byteArr))
			}
		}

		scanInput.ExclusiveStartKey = scanOutput.LastEvaluatedKey
	}
}
