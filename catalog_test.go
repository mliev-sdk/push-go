package mlievpush

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
)

const catalogTestSecret = "catalog-test-secret"

func catalogFixture(t *testing.T) map[string]interface{} {
	t.Helper()
	raw, err := os.ReadFile("testdata/channel.json")
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

// Verify against the wire contract, without using the SDK's signing functions.
func verifyCatalogRequest(t *testing.T, r *http.Request) []byte {
	t.Helper()
	for _, header := range []string{"X-App-Id", "X-Timestamp", "X-Nonce", "X-Signature"} {
		if r.Header.Get(header) == "" {
			t.Errorf("missing %s", header)
		}
	}
	if r.Header.Get("X-App-Id") != "catalog-test" {
		t.Errorf("app ID = %s", r.Header.Get("X-App-Id"))
	}
	if _, err := strconv.ParseInt(r.Header.Get("X-Timestamp"), 10, 64); err != nil {
		t.Errorf("timestamp: %v", err)
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("read request: %v", err)
	}
	canonical := ""
	if r.Method == http.MethodGet {
		if len(body) != 0 {
			t.Errorf("GET body = %s, want empty", body)
		}
	} else {
		var params map[string]interface{}
		if err := json.Unmarshal(body, &params); err != nil {
			t.Errorf("decode request: %v", err)
		}
		encoded, err := json.Marshal(params)
		if err != nil {
			t.Errorf("encode request: %v", err)
		}
		canonical = string(encoded)
	}
	mac := hmac.New(sha256.New, []byte(catalogTestSecret))
	mac.Write([]byte(r.Method + r.URL.Path + canonical + r.Header.Get("X-Timestamp") + r.Header.Get("X-Nonce")))
	if got, want := r.Header.Get("X-Signature"), hex.EncodeToString(mac.Sum(nil)); got != want {
		t.Errorf("signature = %s, want %s (query must be excluded)", got, want)
	}
	return body
}

func writeCatalogResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "message": "success", "data": data})
}

func TestListChannelsQueryAndSignature(t *testing.T) {
	for _, test := range []struct {
		name string
		req  *ListChannelsRequest
		want url.Values
	}{
		{"defaults", nil, url.Values{}},
		{"zero fields", &ListChannelsRequest{}, url.Values{}},
		{"type and pagination", &ListChannelsRequest{Type: "sms", Page: 2, PageSize: 5}, url.Values{"type": {"sms"}, "page": {"2"}, "page_size": {"5"}}},
		{"encoded type", &ListChannelsRequest{Type: "邮件 &+/?"}, url.Values{"type": {"邮件 &+/?"}}},
		{"server validates values", &ListChannelsRequest{Page: -1, PageSize: 101}, url.Values{"page": {"-1"}, "page_size": {"101"}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			item := catalogFixture(t)
			delete(item, "template")
			delete(item, "signature_required")
			delete(item, "signature_names")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				verifyCatalogRequest(t, r)
				if r.Method != "GET" || r.URL.Path != "/api/v1/channels" || !reflect.DeepEqual(r.URL.Query(), test.want) {
					t.Errorf("request = %s %s, want query %v", r.Method, r.URL, test.want)
				}
				writeCatalogResponse(w, map[string]interface{}{"items": []interface{}{item}, "total": 21, "page": 2, "size": 5})
			}))
			defer server.Close()
			result, err := NewClient(server.URL, "catalog-test", catalogTestSecret).ListChannels(context.Background(), test.req)
			if err != nil {
				t.Fatal(err)
			}
			if result.Total != 21 || result.Page != 2 || result.Size != 5 || len(result.Items) != 1 {
				t.Fatalf("list = %+v", result)
			}
			channel := result.Items[0]
			if channel.ID != 42 || channel.Name != "验证码通道" || channel.Type != "sms" || channel.MessageTemplateID != 7 || channel.TemplateName != "登录验证码" || channel.Readiness.State != ChannelReadinessReady || channel.Readiness.BlockerCodes == nil {
				t.Fatalf("channel = %+v", channel)
			}
		})
	}
}

func TestCatalogEmptyListAndNullableDetail(t *testing.T) {
	for _, scenario := range []string{"empty list", "ready", "degraded", "missing template", "invalid variables", "no variables or signature"} {
		t.Run(scenario, func(t *testing.T) {
			data := catalogFixture(t)
			template := data["template"].(map[string]interface{})
			state := ChannelReadinessReady
			switch scenario {
			case "degraded":
				state = ChannelReadinessDegraded
				data["readiness"] = map[string]interface{}{"state": state, "blocker_codes": []string{"PROVIDER_ACCOUNT_UNAVAILABLE"}}
			case "missing template", "invalid variables":
				state = ChannelReadinessBlocked
				code := "MESSAGE_TEMPLATE_VARIABLES_INVALID"
				if scenario == "missing template" {
					data["template"] = nil
					code = "MESSAGE_TEMPLATE_MISSING"
				} else {
					template["variables"] = nil
				}
				data["readiness"] = map[string]interface{}{"state": state, "blocker_codes": []string{code}}
				data["signature_required"], data["signature_names"] = false, []string{}
			case "no variables or signature":
				template["variables"] = []string{}
				data["signature_required"], data["signature_names"] = false, []string{}
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				verifyCatalogRequest(t, r)
				if scenario == "empty list" {
					writeCatalogResponse(w, map[string]interface{}{"items": []interface{}{}, "total": 0, "page": 1, "size": 20})
					return
				}
				if r.Method != "GET" || r.URL.Path != "/api/v1/channels/42" || r.URL.RawQuery != "" {
					t.Errorf("detail request = %s %s", r.Method, r.URL)
				}
				writeCatalogResponse(w, data)
			}))
			defer server.Close()
			client := NewClient(server.URL, "catalog-test", catalogTestSecret)
			if scenario == "empty list" {
				result, err := client.ListChannels(context.Background(), nil)
				if err != nil || result.Items == nil || len(result.Items) != 0 || result.Total != 0 {
					t.Fatalf("empty list = %+v, err=%v", result, err)
				}
				return
			}
			result, err := client.GetChannel(context.Background(), 42)
			if err != nil {
				t.Fatal(err)
			}
			// Round-trip through JSON verifies every documented field and null vs [].
			raw, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			var roundTrip map[string]interface{}
			if err := json.Unmarshal(raw, &roundTrip); err != nil {
				t.Fatal(err)
			}
			expectedRaw, _ := json.Marshal(data)
			var expected map[string]interface{}
			json.Unmarshal(expectedRaw, &expected)
			if !reflect.DeepEqual(roundTrip, expected) || result.Readiness.State != state {
				t.Fatalf("detail = %s, want %s", raw, expectedRaw)
			}
		})
	}
}

func TestCatalogAPIErrorsAndInvalidJSON(t *testing.T) {
	for _, test := range []struct {
		status int
		body   string
		code   int
	}{
		{400, `{"code":400,"message":"invalid query"}`, 400},
		{404, `{"code":404,"message":"channel not found"}`, 404},
		{500, `{"code":500,"message":"query failed"}`, 500},
		{200, `{"code":20003,"message":"invalid signature"}`, 20003},
		{200, `{"code":30001,"message":"rate limited"}`, 30001},
		{502, `not JSON`, 0},
	} {
		t.Run(strconv.Itoa(test.status)+"/"+strconv.Itoa(test.code), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				verifyCatalogRequest(t, r)
				w.WriteHeader(test.status)
				io.WriteString(w, test.body)
			}))
			defer server.Close()
			client := NewClient(server.URL, "catalog-test", catalogTestSecret)
			_, listErr := client.ListChannels(context.Background(), nil)
			_, detailErr := client.GetChannel(context.Background(), 42)
			for _, err := range []error{listErr, detailErr} {
				if err == nil {
					t.Fatal("expected error")
				}
				var apiErr *APIError
				if test.code == 0 {
					if errors.As(err, &apiErr) {
						t.Fatalf("invalid JSON reported as API error: %v", err)
					}
				} else if !errors.As(err, &apiErr) || apiErr.Code != test.code || apiErr.Message == "" {
					t.Fatalf("error = %v, want API code %d", err, test.code)
				}
			}
		})
	}
}

type catalogRoundTripper func(*http.Request) (*http.Response, error)

func (f catalogRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCatalogNetworkErrorsAndCancellation(t *testing.T) {
	failure := errors.New("test network failure")
	client := NewClient("http://example.invalid", "catalog-test", catalogTestSecret, WithHTTPClient(&http.Client{
		Transport: catalogRoundTripper(func(*http.Request) (*http.Response, error) { return nil, failure }),
	}))
	_, listErr := client.ListChannels(context.Background(), nil)
	_, detailErr := client.GetChannel(context.Background(), 42)
	for _, err := range []error{listErr, detailErr} {
		if !errors.Is(err, failure) {
			t.Fatalf("network error not preserved: %v", err)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("canceled request reached server") }))
	defer server.Close()
	client = NewClient(server.URL, "catalog-test", catalogTestSecret)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, listErr = client.ListChannels(ctx, nil)
	_, detailErr = client.GetChannel(ctx, 42)
	for _, err := range []error{listErr, detailErr} {
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation not preserved: %v", err)
		}
	}
}

func TestCatalogSelectionCanBeSent(t *testing.T) {
	fixture := catalogFixture(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := verifyCatalogRequest(t, r)
		requests.Add(1)
		switch r.URL.Path {
		case "/api/v1/channels":
			writeCatalogResponse(w, map[string]interface{}{"items": []interface{}{fixture}, "total": 1, "page": 1, "size": 20})
		case "/api/v1/channels/42":
			writeCatalogResponse(w, fixture)
		case "/api/v1/messages", "/api/v1/messages/batch":
			var request map[string]interface{}
			if err := json.Unmarshal(body, &request); err != nil {
				t.Errorf("decode send body: %v", err)
			}
			if r.Method != "POST" || request["channel_id"] != float64(42) || request["signature_name"] != "验证码" || !reflect.DeepEqual(request["template_params"], map[string]interface{}{"code": "123456", "expire": "5"}) {
				t.Errorf("unexpected send: %s", body)
			}
			writeCatalogResponse(w, map[string]interface{}{"task_id": "catalog-task", "batch_id": "catalog-batch", "status": "pending"})
		case "/api/v1/messages/catalog-task":
			writeCatalogResponse(w, map[string]interface{}{"task_id": "catalog-task", "status": "success"})
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	client := NewClient(server.URL, "catalog-test", catalogTestSecret)
	ctx := context.Background()
	page, err := client.ListChannels(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	detail, err := client.GetChannel(ctx, page.Items[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	values := map[string]string{"code": "123456", "expire": "5"}
	params := make(map[string]string)
	for _, variable := range detail.Template.Variables {
		params[variable] = values[variable]
	}
	sent, err := client.SendMessage(ctx, &SendMessageRequest{ChannelID: detail.ID, Receiver: "13800138000", SignatureName: detail.SignatureNames[0], TemplateParams: params})
	if err != nil || sent.TaskID != "catalog-task" {
		t.Fatalf("send = %+v, err=%v", sent, err)
	}
	if _, err := client.SendBatch(ctx, &SendBatchRequest{ChannelID: detail.ID, Receivers: []string{"13800138000"}, SignatureName: detail.SignatureNames[0], TemplateParams: params}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.QueryTask(ctx, sent.TaskID); err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 5 {
		t.Fatalf("made %d requests, want 5 without implicit pagination or sends", requests.Load())
	}
}
