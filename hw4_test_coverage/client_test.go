package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	dataFile    = "dataset.xml"
	accessToken = "kek"

	ErrorBadQuery       = "ErrorBadQuery"
	ErrorBadAccessToken = "Bad AccessToken"
)

var (
	orderFieldAvailable = map[string]struct{}{
		"Name": struct{}{},
		"Id":   struct{}{},
		"Age":  struct{}{},
		"":     struct{}{},
	}
)

func SearchServer(w http.ResponseWriter, r *http.Request) {
	if value, exist := r.Header[textproto.CanonicalMIMEHeaderKey("AccessToken")]; !exist || value[0] != accessToken {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println("Bad Access Token")
		return
	}
	var err error
	searchRequest := SearchRequest{}

	if searchRequest.Limit, err = strconv.Atoi(r.FormValue("limit")); err != nil {
		log.Printf("Set to default-value %#v", searchRequest.Limit)
	}

	if searchRequest.Offset, err = strconv.Atoi(r.FormValue("offset")); err != nil {
		log.Printf("Set to default-value %#v", searchRequest.Offset)
	}

	if searchRequest.OrderBy, err = strconv.Atoi(r.FormValue("order_by")); err != nil {
		log.Printf("Set to default-value %#v", searchRequest.OrderBy)
	}

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, jsonErr := json.Marshal(&SearchErrorResponse{Error: err.Error()})
		if jsonErr != nil {
			log.Printf("Marshal error: %#v", jsonErr)
		}
		w.Write(data)
		return
	}

	searchRequest.OrderField = r.FormValue("order_field")
	if _, exist := orderFieldAvailable[searchRequest.OrderField]; !exist {
		w.WriteHeader(http.StatusBadRequest)
		data, jsonErr := json.Marshal(&SearchErrorResponse{Error: "ErrorBadOrderField"})
		if jsonErr != nil {
			log.Printf("Marshal error: %#v", jsonErr)
		}
		w.Write(data)
		return
	}
	searchRequest.Query = r.FormValue("query")

	users, err := (XMLUserReader{}).Read(dataFile, searchRequest.Limit, searchRequest.Query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		data, jsonErr := json.Marshal(&SearchErrorResponse{Error: err.Error()})
		if jsonErr != nil {
			log.Printf("Marshal error: %#v", jsonErr)
			return
		}
		w.Write(data)
		return
	}

	log.Printf("Seatch Request %#v\n", searchRequest)
	data, err := json.Marshal(users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data, jsonErr := json.Marshal(&SearchErrorResponse{Error: err.Error()})
		if jsonErr != nil {
			log.Printf("Marshal error: %#v", jsonErr)
			return
		}
		w.Write(data)
		return
	}
	w.Write(data)

}

type UsersReader interface {
	Read(s string, n int, query string) ([]User, error)
}

type XMLUserReader struct{}

type RowXML struct {
	ID         int    `xml:"id"`
	FirstName  string `xml:"first_name"`
	SecondName string `xml:"last_name"`
	Age        int    `xml:"age"`
	About      string `xml:"about"`
	Gender     string `xml:"gender"`
}

func RowXMLToUser(row *RowXML) *User {
	return &User{
		Id:     row.ID,
		Name:   row.FirstName + " " + row.SecondName,
		Age:    row.Age,
		About:  row.About,
		Gender: row.Gender,
	}
}

func (XMLUserReader) Read(s string, n int, query string) ([]User, error) {
	file, err := os.Open(s)
	if err != nil {
		return nil, fmt.Errorf("Can't open file %s for read", s)
	}
	defer file.Close()
	decoder := xml.NewDecoder(file)
	users := make([]User, 0, n)
	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				return users, nil
			}
			return nil, err
		}
		switch e := t.(type) {
		case xml.StartElement:
			if e.Name.Local == "row" {
				row := RowXML{}

				err := decoder.DecodeElement(&row, &e)
				if err != nil {
					return nil, err
				}
				//checkedUserCount++
				user := RowXMLToUser(&row)

				if query == "" || user.Name == query || user.About == query {
					users = append(users, *user)
					log.Printf("Read %#v\n", row)
				}
			}
		}

		if len(users) == n {
			return users, nil
		}
	}
}

type TestCase struct {
	sClient *SearchClient
	sReq    *SearchRequest
	sResp   *SearchResponse
	err     string
}

func TestFindUserBadAccessToken(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{
			AccessToken: "Bad Access Token",
		},
		sReq: &SearchRequest{},
		err:  ErrorBadAccessToken,
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}

}

func TestFindUserBadOrderField(t *testing.T) {

	testCase := TestCase{
		sClient: &SearchClient{
			AccessToken: accessToken,
		},
		sReq: &SearchRequest{
			OrderField: "Incorrect",
		},
		err: "OrderFeld Incorrect invalid",
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}

}

func TimeoutHandle(w http.ResponseWriter, r *http.Request) {
	<-time.After(time.Second * 2)
	w.Write([]byte("ping"))
}

func TestFindUserTimeOut(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "timeout for limit=1&offset=0&order_by=0&order_field=&query=",
	}
	ts := httptest.NewServer(http.HandlerFunc(TimeoutHandle))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}

func TestFindUserNilUrl(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "unknown error Get ?limit=1&offset=0&order_by=0&order_field=&query=: unsupported protocol scheme \"\"",
	}
	ts := httptest.NewServer(http.HandlerFunc(TimeoutHandle))
	defer ts.Close()
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}
func TestFindUserBadLimit(t *testing.T) {

	testCase := TestCase{
		sClient: &SearchClient{
			AccessToken: accessToken,
		},
		sReq: &SearchRequest{
			Limit: -1,
		},
		err: "limit must be > 0",
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}

func FatalErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
}
func TestFindUserStatusInternalServerError(t *testing.T) {

	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "SearchServer fatal error",
	}

	ts := httptest.NewServer(http.HandlerFunc(FatalErrorHandler))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}

}

func BadRequestBadErorrJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"Error\":\""))
}

func TestFindUserBadRequestBadErrorJSON(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "cant unpack error json: unexpected end of JSON input",
	}

	ts := httptest.NewServer(http.HandlerFunc(BadRequestBadErorrJSONHandler))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}
func BadRequestUnknownErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte("{\"Error\":\"Unknown error\"}"))
}
func TestFindUserBadRequestUnknownError(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "unknown bad request error: Unknown error",
	}

	ts := httptest.NewServer(http.HandlerFunc(BadRequestUnknownErrorHandler))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}
func BadUnpackBodyJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Broken json"))
}
func TestFindUserBadUnpackBodyJSON(t *testing.T) {
	testCase := TestCase{
		sClient: &SearchClient{},
		sReq:    &SearchRequest{},
		err:     "cant unpack result json: invalid character 'B' looking for beginning of value",
	}

	ts := httptest.NewServer(http.HandlerFunc(BadUnpackBodyJSONHandler))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}
}

func TestFindUserBadOffset(t *testing.T) {

	testCase := TestCase{
		sClient: &SearchClient{
			AccessToken: accessToken,
		},
		sReq: &SearchRequest{
			Offset: -1,
		},
		err: "offset must be > 0",
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()
	testCase.sClient.URL = ts.URL
	_, err := testCase.sClient.FindUsers(*testCase.sReq)

	if err == nil {
		t.Error("Expected error: ", testCase.err, "\n")
		t.FailNow()
	}
	if err.Error() != testCase.err {
		t.Error("Unexpected error: ", err.Error(), "\n")
		t.FailNow()
	}

}

func initTestCases(url string) []TestCase {
	return []TestCase{
		TestCase{
			sClient: &SearchClient{
				AccessToken: accessToken,
				URL:         url,
			},
			sReq: &SearchRequest{
				Limit:      2,
				Offset:     0,
				OrderBy:    0,
				OrderField: "",
				Query:      "Boyd Wolf",
			},
			sResp: &SearchResponse{
				Users: []User{
					User{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
			},
		},
		// Limit > 25
		TestCase{
			sClient: &SearchClient{
				AccessToken: accessToken,
				URL:         url,
			},
			sReq: &SearchRequest{
				Limit:      26,
				Offset:     0,
				OrderBy:    0,
				OrderField: "",
				Query:      "Boyd Wolf",
			},
			sResp: &SearchResponse{
				Users: []User{
					User{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
			},
		},
		TestCase{
			sClient: &SearchClient{
				AccessToken: accessToken,
				URL:         url,
			},
			sReq: &SearchRequest{
				Limit:      1,
				Offset:     0,
				OrderBy:    0,
				OrderField: "",
				Query:      "",
			},
			sResp: &SearchResponse{
				Users: []User{
					User{
						Id:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: true,
			},
		},
	}
}

func TestFindUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	tc := initTestCases(ts.URL)

	for testNum, test := range tc {
		resp, err := test.sClient.FindUsers(*test.sReq)
		if err == nil && test.err != "" {
			t.Errorf("[%d] expect error: %s\n", testNum, test.err)
			t.FailNow()
		}
		if err != nil && err.Error() != test.err {
			t.Errorf("[%d] unexpected error: %#v", testNum, err.Error())
			t.FailNow()
		}
		if !reflect.DeepEqual(resp, test.sResp) {
			t.Errorf("[%d] unexpected result.\nExpected:\n%#v\nGet:\n%#v\n", testNum, test.sResp, resp)
			t.FailNow()
		}
	}

	ts.Close()
}

// код писать тут
