package eventstore

import (
  "testing"
  "net/http"
  "bytes"
  "io"
  "errors"
  "container/list"
)

func TestNewClient(t *testing.T) {

  tables := []struct {
    baseUrl string
    userName string
    password string
    expectedPass bool
  } {
    {"http://eventstore.hostname:2113", "myuser", "mypass", true},
    {"://eventstore.hostname:2113", "myuser", "mypass", false},
    {"http://eventstore.hostname:2113", "", "", true},
  }

  for _, table := range tables {
  
    client, err := NewClient(table.baseUrl, table.userName, table.password)

    if table.expectedPass == true {
      if err != nil {
        t.Errorf("Unexpected error (%s) raised for inputs %v", err.Error(), table)
        continue
      }

      if client == nil {
        t.Errorf("Client object not created (nil returned) for inputs %v", table)
      }
    } else {
      if err == nil {
        t.Errorf("Expected Client creation to raise error for inputs %v", table)
        continue
      }

      if client != nil {
        t.Errorf("Expected Client creation to fail for inputs %v", table)
      }
    }
  }

}

func TestMakeRequest(t *testing.T) {

  client, _ := NewClient("http://eventstore.hostname:2113", "myuser", "mypass")

  httpClient := NewMockHttpClient()
  client.httpClient = httpClient

  tables := []struct {
    method     string
    body       string
    statusCode int
    status     string
    err        error
  } {
    {"GET",  `{"data": {}}`, 200, "200 OK",        nil},
    {"POST", `{"data": {}}`, 200, "200 OK",        nil},
    {"GET",  `{"data": {}}`, 200, "200 OK",        errors.New("Something went wrong")},
    {"GET",  `{"data": {}}`, 401, "401 Forbidden", nil},
    {"GET",  `{"data": {}}`, 404, "404 Not Found", nil},
  }

  for _, table := range tables {

    httpClient.resetResponses()
    httpClient.addHttpClientResponse(table.body, table.statusCode, table.status, table.err)

    data := map[string]interface{}{}
    err := client.makeRequest(table.method, "/mypath", nil, &data)

    if (table.err != nil || table.statusCode != 200) && err == nil {
      t.Errorf("Expected error not raised for inputs %v", table)
      continue
    } else if table.err == nil && table.statusCode == 200 && err != nil {
      t.Errorf("Unexpected error (%s) raised for inputs %v (%s, %d)", err.Error(), table, err.Error(), table.statusCode)
    }
  }
}

type MockHttpClient struct {
  responses *list.List
  errs *list.List
}

func NewMockHttpClient() (*MockHttpClient) {
  return &MockHttpClient{
    responses: list.New(),
    errs: list.New(),
  }
}

func (client *MockHttpClient) resetResponses() {
  client.responses = list.New()
  client.errs = list.New()
}

func (client *MockHttpClient) Do(*http.Request) (*http.Response, error) {
  err := client.errs.Front()
  client.errs.Remove(err)

  respElem := client.responses.Front()
  resp := respElem.Value.(http.Response)
  client.responses.Remove(respElem)

  if err.Value != nil {
    return nil, err.Value.(error)
  } else {
    return &resp, nil
  }
}

func (client *MockHttpClient) addHttpClientResponse(body string, statusCode int, status string, err error) {
  response := http.Response{
    StatusCode: statusCode,
    Status: status,
    Body: io.NopCloser(bytes.NewReader([]byte(body))),
  }

  client.responses.PushBack(response)
  client.errs.PushBack(err)
}

