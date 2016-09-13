package main_test

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testAuthHandler struct {
	ClientID     string
	ClientSecret string
	token        string
	requests     chan *http.Request
}

func (a *testAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer GinkgoRecover()
	a.requests <- r
	Expect(r.FormValue("client_id")).To(Equal(a.ClientID))
	Expect(r.FormValue("grant_type")).To(Equal("client_credentials"))
	u, p, ok := r.BasicAuth()
	Expect(u).To(Equal(a.ClientID))
	Expect(p).To(Equal(a.ClientSecret))
	Expect(ok).To(BeTrue())
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(map[string]string{
		"token_type":   "Bearer",
		"access_token": a.token,
	})
	Expect(err).ToNot(HaveOccurred())
	w.Write(b)
}
