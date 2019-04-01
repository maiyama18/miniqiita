package miniqiita

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

func TestClient_GetUserItems(t *testing.T) {
	tt := []struct {
		name string

		inputUserID  string
		inputPage    int
		inputPerPage int

		mockResponseHeaderFile string
		mockResponseBodyFile   string

		expectedMethod      string
		expectedRequestPath string
		expectedRawQuery    string
		expectedItems       []*Item
		expectedErrMessage  string
	}{}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.Method != tc.expectedMethod {
					t.Fatalf("request method wrong. want=%s, got=%s", tc.expectedMethod, req.Method)
				}
				if req.URL.Path != tc.expectedRequestPath {
					t.Fatalf("request path wrong. want=%s, got=%s", tc.expectedRequestPath, req.URL.Path)
				}
				if req.URL.RawQuery != tc.expectedRawQuery {
					t.Fatalf("request query wrong. want=%s, got=%s", tc.expectedRawQuery, req.URL.RawQuery)
				}

				headerBytes, err := ioutil.ReadFile(tc.mockResponseHeaderFile)
				if err != nil {
					t.Fatalf("failed to read header '%s': %s", tc.mockResponseHeaderFile, err.Error())
				}
				firstLine := strings.Split(string(headerBytes), "\n")[0]
				statusCode, err := strconv.Atoi(strings.Fields(firstLine)[1])
				if err != nil {
					t.Fatalf("failed to extract status code from header: %s", err.Error())
				}
				w.WriteHeader(statusCode)

				bodyBytes, err := ioutil.ReadFile(tc.mockResponseBodyFile)
				if err != nil {
					t.Fatalf("failed to read body '%s': %s", tc.mockResponseBodyFile, err.Error())
				}
				w.Write(bodyBytes)
			}))
			defer server.Close()

			serverURL, err := url.Parse(server.URL)
			if err != nil {
				t.Fatalf("failed to get mock server URL: %s", err.Error())
			}

			cli := &Client{
				BaseURL:    serverURL,
				HTTPClient: server.Client(),
				Logger:     nil,
			}

			items, err := cli.GetUserItems(context.Background(), tc.inputUserID, tc.inputPage, tc.inputPerPage)
			if tc.expectedErrMessage == "" {
				if err != nil {
					t.Fatalf("response error should be nil. got=%s", err.Error())
				}

				if len(items) != len(tc.expectedItems) {
					t.Fatalf("response items wrong. want=%+v, got=%+v", tc.expectedItems, items)
				}
				for i, expected := range tc.expectedItems {
					actual := items[i]
					if actual.ID != expected.ID || actual.Title != expected.Title || actual.LikesCount != actual.LikesCount {
						t.Fatalf("response items wrong. want=%+v, got=%+v", tc.expectedItems, items)
					}
				}
			} else {
				if err == nil {
					t.Fatalf("response error should not be non-nil. got=nil")
				}

				if !strings.Contains(err.Error(), tc.expectedErrMessage) {
					t.Fatalf("response error message wrong. '%s' is expected to contain '%s'", err.Error(), tc.expectedErrMessage)
				}
			}
		})
	}
}
