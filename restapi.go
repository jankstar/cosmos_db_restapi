/*
package cosmos_db_restapi
Golang Package for the cosmos db rest api
*/
package cosmos_db_restapi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
)

/*
TQuerry structure for querry call:

	{
		"query":"SELECT * FROM c WHERE c.name = @name",
		"parameters": [{
			"name": "@name",
			"value": "Julian"
		}]
	}
*/
type TParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type TQuerry struct {
	Query      string       `json:"query"`
	Parameters []TParameter `json:"parameters"`
}

type TBody struct {
	Code      string        `json:"code"`      //error code from cosmos_db
	Message   string        `json:"message"`   //error message
	Rid       string        `json:"_rid"`      //resource ID
	Documents []interface{} `json:"Documents"` //array of data
	Count     uint          `json:"_count"`    //counts of documents
}

/*
GetAuthorizationTokenUsingMasterKey
function for generating access token

https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources
*/
func GetAuthorizationTokenUsingMasterKey(
	verb string,
	resourceType string,
	resourceId string,
	date string,
	masterKey string) string {

	key, _ := base64.StdEncoding.DecodeString(masterKey)

	text := strings.ToLower(verb) + "\n" +
		strings.ToLower(resourceType) + "\n" +
		//		strings.ToLower(resourceId) + "\n" +
		resourceId + "\n" +
		strings.ToLower(date) + "\n" +
		"" + "\n"

	body := []byte(text)

	hash := hmac.New(sha256.New, []byte(key))
	_, _ = hash.Write(body)
	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	MasterToken := "master"

	TokenVersion := "1.0"

	return url.QueryEscape("type=" + MasterToken + "&ver=" + TokenVersion + "&sig=" + signature)
}

/*
ExecuteQuerry - execute a query as rest api

parameters:

	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition key else ""
	max_item_count - optional max item count else 0
	querry - like TQuerry

returns:

	Status - response status i.e. 200 ok
	Body - response body as string
	Continuation - the Continuation-token if there are more items to read
*/
func ExecuteQuerry(endpoint_uri string, master_key string, database string, container string, partitionkey string, max_item_count int, continuation string, query TQuerry) (Status string, Body string, Continuation string) {

	date_str := strings.ToLower(time.Now().UTC().Format(http.TimeFormat))

	resource_link := strings.ToLower("dbs/" + database + "/colls/" + container)

	autorization_str := GetAuthorizationTokenUsingMasterKey("POST", "docs",
		resource_link, date_str, master_key)

	url := endpoint_uri + resource_link + "/docs"

	querry_json, _ := json.Marshal(query)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(querry_json))

	req.Header.Set("Accept", "*/*")
	req.Header.Set("x-ms-documentdb-isquery", "True")
	req.Header.Set("x-ms-documentdb-query-enablecrosspartition", "True")
	req.Header.Set("authorization", autorization_str)

	req.Header.Set("x-ms-version", "2020-11-05")
	req.Header.Set("x-ms-date", date_str)
	req.Header.Set("Content-Type", "application/query+json")
	if partitionkey != "" {
		req.Header.Set("x-ms-documentdb-partitionkey", "[ "+"\""+partitionkey+"\""+" ]")
	} else {
		req.Header.Set("x-ms-documentdb-query-enablecrosspartition", "True")
	}

	if max_item_count > 0 {
		req.Header.Set("x-ms-max-item-count", strconv.Itoa(max_item_count))
	}

	if continuation != "" {
		req.Header.Set("x-ms-continuation", continuation)
	}

	//req.Header.Set("x-ms-documentdb-populatequerymetrics", "True")

	http_client := &http.Client{}
	res, _ := http_client.Do(req)

	Continuation = res.Header.Get("x-ms-continuation")

	res_body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return res.Status, string(res_body), Continuation
}

/*
GetDocumentByID - get an object by ID via rest api

parameters:

	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	id - id of the item

returns:

	Status - response status i.e. 200 ok
	Body - response body as string
*/
func GetDocumentByID(endpoint_uri string, master_key string, database string, container string, partitionkey string, id string) (Status string, Body string) {

	date_str := strings.ToLower(time.Now().UTC().Format(http.TimeFormat))

	resource_link := strings.ToLower("dbs/"+database+"/colls/"+container+"/docs/") + id

	autorization_str := GetAuthorizationTokenUsingMasterKey("GET", "docs", resource_link, date_str, master_key)

	url := endpoint_uri + resource_link

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Set("Accept", "*/*")

	req.Header.Set("authorization", autorization_str)

	req.Header.Set("x-ms-version", "2020-11-05")
	req.Header.Set("x-ms-date", date_str)
	req.Header.Set("Content-Type", "application/json")
	if partitionkey != "" {
		req.Header.Set("x-ms-documentdb-partitionkey", "[ "+"\""+partitionkey+"\""+" ]")
	}

	//req.Header.Set("x-ms-documentdb-populatequerymetrics", "True")

	http_client := &http.Client{}
	res, _ := http_client.Do(req)

	res_body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return res.Status, string(res_body)
}

/*
CreateDocument - create or rewrite an object by ID via rest api

parameters:
	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	upset -- optional boolean, true: create or update, if exist
	data - json data sa sting of the item

returns:
	Status - response status i.e. 201 Created
	Body - response body as string
*/

func CreateDocument(endpoint_uri string, master_key string, database string, container string, partitionkey string, upset bool, data string) (Status string, Body string) {

	date_str := strings.ToLower(time.Now().UTC().Format(http.TimeFormat))

	resource_link := strings.ToLower("dbs/" + database + "/colls/" + container)

	autorization_str := GetAuthorizationTokenUsingMasterKey("POST", "docs", resource_link, date_str, master_key)

	url := endpoint_uri + resource_link + "/docs"

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))

	req.Header.Set("Accept", "*/*")

	req.Header.Set("authorization", autorization_str)

	req.Header.Set("x-ms-version", "2020-11-05")
	req.Header.Set("x-ms-date", date_str)
	req.Header.Set("Content-Type", "application/json")
	if upset == true {
		req.Header.Set("x-ms-documentdb-is-upsert", "True") //create or update if exist
	}
	if partitionkey != "" {
		req.Header.Set("x-ms-documentdb-partitionkey", "[ "+"\""+partitionkey+"\""+" ]")
	}

	//req.Header.Set("x-ms-documentdb-populatequerymetrics", "True")

	http_client := &http.Client{}
	res, _ := http_client.Do(req)

	res_body, _ := ioutil.ReadAll(res.Body)

	return res.Status, string(res_body)
}

/*
DeleteDocumentByID - delete an object by ID via rest api

parameters:

	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	id - id of the item

returns:

	Status - response status i.e. 204 No Content
	Body - response body as string i.e. ""
*/
func DeleteDocumentByID(endpoint_uri string, master_key string, database string, container string, partitionkey string, id string) (Status string, Body string) {

	date_str := strings.ToLower(time.Now().UTC().Format(http.TimeFormat))

	resource_link := strings.ToLower("dbs/"+database+"/colls/"+container+"/docs/") + id

	autorization_str := GetAuthorizationTokenUsingMasterKey("DELETE", "docs", resource_link, date_str, master_key)

	url := endpoint_uri + resource_link

	req, _ := http.NewRequest("DELETE", url, nil)

	req.Header.Set("Accept", "*/*")

	req.Header.Set("authorization", autorization_str)

	req.Header.Set("x-ms-version", "2020-11-05")
	req.Header.Set("x-ms-date", date_str)
	req.Header.Set("Content-Type", "application/json")
	if partitionkey != "" {
		req.Header.Set("x-ms-documentdb-partitionkey", "[ "+"\""+partitionkey+"\""+" ]")
	}

	//req.Header.Set("x-ms-documentdb-populatequerymetrics", "True")

	http_client := &http.Client{}
	res, _ := http_client.Do(req)

	res_body, _ := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	return res.Status, string(res_body)
}

// test()
func test() {
	//get the "endpoint" and master-key from the .env file
	godotenv.Load(".env")
	endpoint := os.Getenv("ENDPOINT_URI")
	key := os.Getenv("MASTER_KEY")

	var querry = TQuerry{
		Query: "SELECT * FROM c WHERE c.word = @word1 OR c.word = @word2 ",
		Parameters: []TParameter{
			{
				Name:  "@word1",
				Value: "Zwerg",
			},
			{
				Name:  "@word2",
				Value: "Nase",
			}},
	}

	req_continuation := ""
	steps := 0
	for req_continuation != "" || steps == 0 {
		steps += 1
		fmt.Println("Step:", steps)

		res_status, res_body, res_continuation := ExecuteQuerry(
			endpoint, key,
			"lerneria-express", "dictionary", "", //no patition key
			3, req_continuation, //only 3 documents per request
			querry)
		fmt.Println("Status: " + res_status)
		//last "continuation" header becomes new request value
		req_continuation = res_continuation

		var MyBody TBody
		_ = json.Unmarshal([]byte(res_body), &MyBody)
		fmt.Println("Count: " + strconv.Itoa(int(MyBody.Count)))
		fmt.Println("continuation:", res_continuation)
		fmt.Println("Documents:")
		for _, doc := range MyBody.Documents {

			//mapping the interface{} element to struc
			type Dic struct {
				Etag      string `mapstructure:"_etag"`
				Rid       string `mapstructure:"_rid"`
				ID        string `mapstructure:"id"`
				Word      string `mapstructure:"word"`
				Snippet   string `mapstructure:"snippet"`
				CreatedAt string `mapstructure:"created_at"`
			}
			var MyDic Dic
			mapstructure.Decode(doc, &MyDic)

			fmt.Println(MyDic)
		}
	}

}