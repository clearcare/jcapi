package jcapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"strings"
)

const (
	responseSize = 256 * 1024
	stdUrlBase   = "https://console.jumpcloud.com/api"
)

type JCOp uint8

const (
	Read   JCOp = 1
	Insert JCOp = 2
	Update JCOp = 3
	Delete JCOp = 4
	List   JCOp = 5
)

type JCAPI struct {
	ApiKey  string
	UrlBase string
}

const (
	searchLimit        int = 1000
	searchSkipInterval int = 1000
)

type JCError interface {
	Error() string
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

func NewJCAPI(apiKey string, urlBase string) JCAPI {
	return JCAPI{
		ApiKey:  apiKey,
		UrlBase: urlBase,
	}
}

func buildJSONStringArray(field string, s []string) string {
	returnVal := "["

	if s != nil {
		afterFirst := false

		for _, val := range s {
			if afterFirst {
				returnVal += ","
			}

			returnVal += "\"" + val + "\""

			afterFirst = true
		}
	}
	returnVal += "]"

	return "\"" + field + "\":" + returnVal
}

func buildJSONKeyValuePair(key, value string) string {
	return "\"" + key + "\":\"" + value + "\""
}

func buildJSONKeyValueBoolPair(key string, value bool) string {
	if value == true {
		return "\"" + key + "\":\"true\""
	} else {
		return "\"" + key + "\":\"false\""
	}

}

func getTimeString() string {
	t := time.Now()

	return t.Format(time.RFC3339)
}

func (jc JCAPI) emailFilter(email string) []byte {

	//
	// Ideally, this would be generalized to take a map[string]string,
	// that doesn't elicit the correct JSON output for the JumpCloud
	// filters in json.Marshal()
	//
	return []byte(fmt.Sprintf("{\"filter\": [{\"email\" : \"%s\"}]}", email))
}

//being lazy; copy paste
func (jc JCAPI) hostnameFilter(hostname string) []byte {

	//
	// Ideally, this would be generalized to take a map[string]string,
	// that doesn't elicit the correct JSON output for the JumpCloud
	// filters in json.Marshal()
	//
	return []byte(fmt.Sprintf("{\"filter\": [{\"hostname\" : \"%s\"}]}", hostname))
}

func (jc JCAPI) setHeader(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-api-key", jc.ApiKey)
}

func (jc JCAPI) Post(url string, data []byte) (interface{}, JCError) {
	return jc.Do(MapJCOpToHTTP(Insert), url, data)
}

func (jc JCAPI) Put(url string, data []byte) (interface{}, JCError) {
	return jc.Do(MapJCOpToHTTP(Update), url, data)
}

func (jc JCAPI) Delete(url string) (interface{}, JCError) {
	return jc.Do(MapJCOpToHTTP(Delete), url, nil)
}

func (jc JCAPI) Get(url string) (interface{}, JCError) {
	return jc.Do(MapJCOpToHTTP(Read), url, nil)
}

func (jc JCAPI) List(url string) (interface{}, JCError) {
	return jc.Do(MapJCOpToHTTP(List), url, nil)
}

func (jc JCAPI) Do(op, url string, data []byte) (interface{}, JCError) {
	var returnVal interface{}

	fullUrl := jc.UrlBase + url

	client := &http.Client{}

	req, err := http.NewRequest(op, fullUrl, bytes.NewReader(data))
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not build search request: '%s'", err)
	}

	jc.setHeader(req)

	resp, err := client.Do(req)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: client.Do() failed, err='%s'", err)
	}

	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		return returnVal, fmt.Errorf("JumpCloud HTTP response status='%s'", resp.Status)
	}

	buffer, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not read the response body, err='%s'", err)
	}

	err = json.Unmarshal(buffer, &returnVal)
	if err != nil {
		return returnVal, fmt.Errorf("ERROR: Could not Unmarshal JSON response, err='%s'", err)
	}

	return returnVal, err
}

// Add all the tags of which the user is a part to the JCUser object
func (user *JCUser) AddJCTags(tags []JCTag) {
	for _, tag := range tags {
		for _, systemUser := range tag.SystemUsers {
			if systemUser == user.Id {
				user.Tags = append(user.Tags, tag)
			}
		}
	}
}

// Add all the tags of which the system is a part to the JCSystem object
func (system *JCSystem) AddJCTagsToSystem(tags []JCTag) {
	for _, tag := range tags {
		for _, sys := range tag.Systems {
			if sys == system.Id {
				system.Tags = append(system.Tags, tag)
			}
		}
	}
}

func MapJCOpToHTTP(op JCOp) string {
	var returnVal string

	switch op {
	case Read:
		returnVal = "GET"
	case Insert:
		returnVal = "POST"
	case Update:
		returnVal = "PUT"
	case Delete:
		returnVal = "DELETE"
	case List:
		returnVal = "LIST"
	}

	return returnVal
}

//
// Interface Conversion Helper Functions
//
func (jc JCAPI) extractStringArray(input []interface{}) []string {
	var returnVal []string

	for _, str := range input {
		returnVal = append(returnVal, str.(string))
	}

	return returnVal
}

func getStringOrNil(input interface{}) string {
	returnVal := ""

	switch input.(type) {
	case string:
		returnVal = input.(string)
	}

	return returnVal
}

func getUint16OrNil(input interface{}) uint16 {
	var returnVal uint16

	switch input.(type) {
	case uint16:
		returnVal = input.(uint16)
	}

	return returnVal
}

func GetTrueOrFalse(input interface{}) bool {
	returnVal := false

	switch input.(type) {
	case string:
		temp := strings.ToLower(input.(string))
		returnVal = strings.Contains("true", temp) || strings.Contains("yes", temp) || strings.Contains("1", temp)
		break
	case int:
		returnVal = input.(int) != 0
		break
	case bool:
		returnVal = input.(bool)
		break
	}

	return returnVal
}
