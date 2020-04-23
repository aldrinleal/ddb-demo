package analytics

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/PuerkitoBio/purell"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Type Alias
type RecordMap map[string]string

var timeFields = map[string]bool{
	"dtm": true,
	"stm": true,
}

func getRevDomainFor(urlStr string) (map[string]string, error) {
	normalizedUrlString, err := purell.NormalizeURLString(urlStr, purell.FlagLowercaseScheme|purell.FlagLowercaseHost|purell.FlagUppercaseEscapes)

	if nil != err {
		log.Warnf("Oops: %s", err)

		return map[string]string{}, err
	}

	parsedUrl, err := url.Parse(normalizedUrlString)

	if nil != err {
		log.Warnf("Oops: %s", err)

		return map[string]string{}, err
	}

	etldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(parsedUrl.Host)

	if nil != err {
		log.Warnf("Oops: %s", err)

		return map[string]string{}, err
	}

	elements := strings.Split(etldPlusOne, ".")

	for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
		elements[i], elements[j] = elements[j], elements[i]
	}

	reversed_domain := strings.Join(elements, ":")

	domain_md5 := md5.Sum([]byte(reversed_domain))

	return map[string]string{
		"domain_md5":     hex.EncodeToString(domain_md5[:]) + "-" + reversed_domain,
		"domain":         reversed_domain,
		"hostname":       parsedUrl.Hostname(),
		"normalized_url": normalizedUrlString,
		"request_uri":    parsedUrl.RequestURI(),
	}, nil
}

func (m RecordMap) AsWriteRequest() (*dynamodb.WriteRequest, error) {
	metadata, err := getRevDomainFor(m["url"])

	if nil != err {
		log.Warnf("Oops: %s", err)

		return nil, err
	}

	attrMap := map[string]*dynamodb.AttributeValue{}

	for k, v := range m {
		// is it a time field?
		if _, ok := timeFields[k]; ok {
			vAsInt64, err := strconv.ParseFloat(v, 64)

			if nil != err {
				log.Warnf("Oops: %s", err)

				continue
			}

			t := time.Unix(int64(vAsInt64/1000), 0)

			vAsUnix := fmt.Sprintf("%d", t.Unix())

			attrMap[k] = &dynamodb.AttributeValue{N: aws.String(vAsUnix)}

			continue
		}

		// otherwise map as string

		attrMap[k] = &dynamodb.AttributeValue{S: aws.String(v)}
	}

	for k, v := range metadata {
		attrMap[k] = &dynamodb.AttributeValue{S: aws.String(v)}
	}

	attrMap["ttm"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", time.Now().Unix()))}
	attrMap["ttl"] = &dynamodb.AttributeValue{N: aws.String(strconv.FormatInt(time.Now().Add(60*24*time.Hour).Unix(), 10))}

	return &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: attrMap,
		},
	}, nil
}

// EventsRecord means a snowplow record - a very generic one
type EventsRecord struct {
	Data []RecordMap `json:"data"`
}
