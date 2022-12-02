# golang_cosmos_db_restapi
golang sql api for cosmos db

```go get github.com/jankstar/golang_cosmos_db_restapi```

# Prerequisites
go version 1.18+

# Function

## GetAuthorizationTokenUsingMasterKey
GetAuthorizationTokenUsingMasterKey function for generating access token
https://docs.microsoft.com/en-us/rest/api/cosmos-db/access-control-on-cosmosdb-resources

### Parameters:
	Verb: "GET"
	Ressourcentyp: "dbs"
	Resource Linkage: "dbs/ToDoList"
	Date: "Thu, 27 Apr 2017 00:51:12 GMT"
	masterKey: "dsZQi3KtZmCv1ljt3VNWNm7sQUF1y5rJfC6kv5JiwvW0EndXdDku/dkKBp8/ufDToSxLzR4y+O/0H/t4bQtVNw=="

### Return:	
	authorization: "typ%3dmaster%26ver%3d1.0%26sig%3dc09PEVJrgp2uQRkr934kFbTqhByc7TVr3OHyqlu%2bc%2bc%2bc%3d"

## ExecuteQuerry
ExecuteQuerry - execute a query as rest api

https://docs.microsoft.com/en-us/rest/api/cosmos-db/querying-cosmosdb-resources-using-the-rest-api
### Parameters:
	endpoint_uri - uri from cosmos db
	master_key - masker key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition key else ""
	max_item_count - optional max item count else 0
	querry - like TQuery

### Returns:
	Status - response status i.e. 200 ok
	Body - response body as string
	Continuation - the Continuation-token if there are more items to read

## CreateDocument
CreateDocument - create or rewrite an object by ID via rest api

https://docs.microsoft.com/en-us/rest/api/cosmos-db/create-a-document
### Parameters:
	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	upset -- optional boolean, true: create or update, if exist
	data - json data as sting of the item

### Returns:
	Status - response status i.e. 201 Created
	Body - response body as string

## GetDocumentByID
GetDocumentByID - get an object by ID via rest api

https://docs.microsoft.com/en-us/rest/api/cosmos-db/get-a-document
### Parameters:
	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	id - id of the item

### Returns:
	Status - response status i.e. 200 ok
	Body - response body as string


## DeleteDocumentByID
DeleteDocumentByID - delete an object by ID via rest api

https://docs.microsoft.com/en-us/rest/api/cosmos-db/delete-a-document
### Parameters:
	endpoint_uri - uri from cosmos db
	master_key - master key from cosmos db
	database - name of database
	container - name of container
	partitionkey - optional partition (if container defined with partion key, it is required)
	id - id of the item

### Returns:
	Status - response status i.e. 204 No Content 
	Body - response body as string i.e. ""
	
## Example 1 - native operations
```
func test() {
	//get the "endpoint" and master-key from the .env file
	godotenv.Load(".env")
	endpoint := os.Getenv("ENDPOINT_URI")
	key := os.Getenv("MASTER_KEY")

	var querry = TQuery{
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
```

## Example 2 - object-like operations
```
func test() {
	//get the "endpoint" and master-key from the .env file
	godotenv.Load(".env")
	endpoint := os.Getenv("ENDPOINT_URI")
	key := os.Getenv("MASTER_KEY")

	var querry = TQuery{
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

	container := ContainerFactory(
		DatabaseFactory(
			endpoint,
			key,
			"lerneria-express"),
		"dictionary",
		"")

	container.OpenQuerry(3, querry) //set query for fetching with 3 docs

	for {
		res_status, res_body := container.Fetch() //fetching data
		fmt.Println("Status: " + res_status)
		if !strings.Contains(res_status, "200") || res_body == "" {
			break //Cancel because error
		}

		var MyBody TBody
		_ = json.Unmarshal([]byte(res_body), &MyBody)
		fmt.Println("Count: " + strconv.Itoa(int(MyBody.Count)))
		fmt.Println("continuation:", container.Continuation)
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
```
