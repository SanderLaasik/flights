package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestMain(m *testing.M) {
	setupConf()
	setupDatabase()
	setupRouter()
	os.Exit(m.Run())
}

func TestAirportsHappyPath(t *testing.T) {

	w := makeRequest(http.MethodGet, "/airports")

	assert.Equal(t, w.Code, 200)
}

func TestFlightsHappyPath(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, 200)
}

func TestFlightsFromMissing(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?to=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.From' Error:Field validation for 'From' failed on the 'required' tag"}`)
}

func TestFlightsToMissing(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.To' Error:Field validation for 'To' failed on the 'required' tag"}`)
}

func TestFlightsDateMissing(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=JFK")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.Date' Error:Field validation for 'Date' failed on the 'required' tag"}`)
}

func TestFlightsFromLong(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAXX&to=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.From' Error:Field validation for 'From' failed on the 'len' tag"}`)
}

func TestFlightsToLong(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=JFKK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.To' Error:Field validation for 'To' failed on the 'len' tag"}`)
}

func TestFlightsFromShort(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LA&to=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.From' Error:Field validation for 'From' failed on the 'len' tag"}`)
}

func TestFlightsToShort(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=JF&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.To' Error:Field validation for 'To' failed on the 'len' tag"}`)
}

func TestFlightsFromNotAlphanum(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=ÜÜÜ&to=JFK&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.From' Error:Field validation for 'From' failed on the 'alphanum' tag"}`)
}

func TestFlightsToNotAlphanum(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=ÜÜÜ&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.To' Error:Field validation for 'To' failed on the 'alphanum' tag"}`)
}

func TestFlightsToNotSameAsFrom(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=LAX&to=LAX&date=2008/01/01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
	assert.Equal(t, w.Body.String(), `{"error":"Key: 'ConnectionParams.To' Error:Field validation for 'To' failed on the 'nefield' tag"}`)
}

func TestFlightsDateInvalidFormat(t *testing.T) {

	w := makeRequest(http.MethodGet, "/connections?from=JFK&to=LAX&date=2008-01-01")

	assert.Equal(t, w.Code, http.StatusBadRequest)
}

func makeRequest(method string, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, nil)
	router.ServeHTTP(w, req)

	return w
}
