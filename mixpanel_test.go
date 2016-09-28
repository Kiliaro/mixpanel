package mixpanel

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

var (
	ts          *httptest.Server
	client      Mixpanel
	LastRequest *http.Request
)

func setup() {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("1\n"))
		LastRequest = r
	}))

	client = New("e3bc4100330c35722740fb8c6f5abddc", ts.URL)
}

func teardown() {
	ts.Close()
}

func decodeURL(url string) string {
	data := strings.Split(url, "data=")[1]
	decoded, _ := base64.StdEncoding.DecodeString(data)
	return string(decoded[:])
}

// examples from https://mixpanel.com/help/reference/http

func TestTrack(t *testing.T) {
	setup()
	defer teardown()

	client.Track("13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
	})

	want := "{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/track"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestPeopleOperations(t *testing.T) {
	setup()
	defer teardown()

	client.Update("13793", &Update{
		Operation: "$set",
		Properties: map[string]interface{}{
			"Address":  "1313 Mockingbird Lane",
			"Birthday": "1948-01-01",
		},
	})

	want := "{\"$distinct_id\":\"13793\",\"$set\":{\"Address\":\"1313 Mockingbird Lane\",\"Birthday\":\"1948-01-01\"},\"$token\":\"e3bc4100330c35722740fb8c6f5abddc\"}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/engage"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestPeopleTrack(t *testing.T) {
	setup()
	defer teardown()

	client.Track("13793", "Signed Up", &Event{
		Properties: map[string]interface{}{
			"Referred By": "Friend",
		},
	})

	want := "{\"event\":\"Signed Up\",\"properties\":{\"Referred By\":\"Friend\",\"distinct_id\":\"13793\",\"token\":\"e3bc4100330c35722740fb8c6f5abddc\"}}"

	if !reflect.DeepEqual(decodeURL(LastRequest.URL.String()), want) {
		t.Errorf("LastRequest.URL returned %+v, want %+v",
			decodeURL(LastRequest.URL.String()), want)
	}

	want = "/track"
	path := LastRequest.URL.Path

	if !reflect.DeepEqual(path, want) {
		t.Errorf("path returned %+v, want %+v",
			path, want)
	}
}

func TestError(t *testing.T) {
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("0\n"))
		LastRequest = r
	}))

	client = New("e3bc4100330c35722740fb8c6f5abddc", ts.URL)

	if err := client.Update("1", &Update{}); err != ErrTrackFailed {
		t.Error("Got bad error for track", err)
	}

	if err := client.Track("1", "name", &Event{}); err != ErrTrackFailed {
		t.Error("Got bad error for track", err)
	}
}

func TestUpdateUnion(t *testing.T) {
	client := NewMock()

	updates := []Update{
		Update{
			Operation:  "$union",
			IP:         "127.0.0.1",
			Properties: map[string]interface{}{"f": []string{"a"}},
		},
		Update{
			Operation:  "$union",
			IP:         "127.0.0.1",
			Properties: map[string]interface{}{"f": []string{"b"}},
		},
		Update{
			Operation:  "$union",
			IP:         "127.0.0.1",
			Properties: map[string]interface{}{"f": []string{"c", "a", "b", "d"}},
		},
	}

	for _, u := range updates {
		if err := client.Update("1", &u); err != nil {
			t.Fatal(err)
		}
	}

	f := client.people("1").Properties["f"]
	if !reflect.DeepEqual(f, []string{"a", "b", "c", "d"}) {
		t.Error("Bad union:", f)
	}
}
