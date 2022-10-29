package cosmos_db_restapi

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
)

func TestExecuteQuerry(t *testing.T) {

	godotenv.Load(".env")

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

	type args struct {
		endpoint_uri   string
		master_key     string
		database       string
		container      string
		partitionkey   string
		max_item_count int
		continuation   string
		querry         TQuerry
	}
	tests := []struct {
		name             string
		args             args
		wantStatus       string
		wantBody         string
		wantContinuation string
	}{
		{
			name: "Test 1",
			args: args{
				endpoint_uri:   os.Getenv("ENDPOINT_URI"),
				master_key:     os.Getenv("MASTER_KEY"),
				database:       os.Getenv("DATABASE"),
				container:      "dictionary",
				partitionkey:   "",
				max_item_count: 0,
				continuation:   "",
				querry:         querry,
			},
			wantStatus:       "200 OK",
			wantBody:         "",
			wantContinuation: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			res_status := ""
			res_body := ""
			res_continuation := ""
			req_continuation := ""
			steps := 0
			for req_continuation != "" || steps == 0 {
				steps += 1
				fmt.Println("Step:", steps)

				res_status, res_body, res_continuation = ExecuteQuerry(
					tt.args.endpoint_uri, tt.args.master_key,
					tt.args.database, tt.args.container, tt.args.partitionkey,
					tt.args.max_item_count, tt.args.continuation,
					tt.args.querry)

				fmt.Println("Status: " + res_status)
				req_continuation = res_continuation

				var MyBody TBody
				_ = json.Unmarshal([]byte(res_body), &MyBody)
				fmt.Println("Count: " + strconv.Itoa(int(MyBody.Count)))
				fmt.Print("continuation:", res_continuation)
				fmt.Println("Documents:")
				for _, doc := range MyBody.Documents {

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

			if res_status != tt.wantStatus {
				t.Errorf("ExecuteQuerry() gotStatus = %v, want %v", res_status, tt.wantStatus)
			}

		})
	}
}

func TestGetDocumentByID(t *testing.T) {

	godotenv.Load(".env")

	type args struct {
		endpoint_uri string
		master_key   string
		database     string
		container    string
		partitionkey string
		id           string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus string
		wantBody   string
	}{
		{
			name: "Test 2",
			args: args{
				endpoint_uri: os.Getenv("ENDPOINT_URI"),
				master_key:   os.Getenv("MASTER_KEY"),
				database:     os.Getenv("DATABASE"),
				container:    "dictionary",
				partitionkey: "Zwerg",
				id:           "Zwerg",
			},
			wantStatus: "200 OK",
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotBody := GetDocumentByID(tt.args.endpoint_uri, tt.args.master_key, tt.args.database, tt.args.container, tt.args.partitionkey, tt.args.id)
			if gotStatus != tt.wantStatus {
				t.Errorf("GetDocumentByID() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			fmt.Println(string(gotBody))
			// if gotBody != tt.wantBody {
			// 	t.Errorf("GetDocumentByID() gotBody = %v, want %v", gotBody, tt.wantBody)
			// }
		})
	}
}

func TestCreateDocument(t *testing.T) {

	godotenv.Load(".env")

	type args struct {
		endpoint_uri string
		master_key   string
		database     string
		container    string
		partitionkey string
		upset        bool
		data         string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus string
		wantBody   string
	}{
		{
			name: "Test 3",
			args: args{
				endpoint_uri: os.Getenv("ENDPOINT_URI"),
				master_key:   os.Getenv("MASTER_KEY"),
				database:     os.Getenv("DATABASE"),
				container:    "user",
				partitionkey: "SuperJoda2",
				upset:        false,
				data:         `{"id": "SuperJoda2","name": "SuperJoda2"}`,
			},
			wantStatus: "201 Created",
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotBody := CreateDocument(tt.args.endpoint_uri, tt.args.master_key, tt.args.database, tt.args.container, tt.args.partitionkey, tt.args.upset, tt.args.data)
			if gotStatus != tt.wantStatus {
				t.Errorf("CreateDocument() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			fmt.Println(string(gotBody))

			// if gotBody != tt.wantBody {
			// 	t.Errorf("CreateDocument() gotBody = %v, want %v", gotBody, tt.wantBody)
			// }
		})
	}
}

func TestDeleteDocumentByID(t *testing.T) {

	godotenv.Load(".env")

	type args struct {
		endpoint_uri string
		master_key   string
		database     string
		container    string
		partitionkey string
		id           string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus string
		wantBody   string
	}{
		{
			name: "Test 4",
			args: args{
				endpoint_uri: os.Getenv("ENDPOINT_URI"),
				master_key:   os.Getenv("MASTER_KEY"),
				database:     os.Getenv("DATABASE"),
				container:    "user",
				partitionkey: "SuperJoda2",
				id:           "SuperJoda2",
			},
			wantStatus: "204 No Content",
			wantBody:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotBody := DeleteDocumentByID(tt.args.endpoint_uri, tt.args.master_key, tt.args.database, tt.args.container, tt.args.partitionkey, tt.args.id)
			if gotStatus != tt.wantStatus {
				t.Errorf("GetDocumentByID() gotStatus = %v, want %v", gotStatus, tt.wantStatus)
			}
			fmt.Println(string(gotBody))
			// if gotBody != tt.wantBody {
			// 	t.Errorf("GetDocumentByID() gotBody = %v, want %v", gotBody, tt.wantBody)
			// }
		})
	}
}
