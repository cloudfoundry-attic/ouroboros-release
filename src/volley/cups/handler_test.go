package cups_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"volley/cups"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {
	It("returns a drain bindings", func() {
		store := &SpyAppIDStore{}
		handler := cups.NewCUPSHandler(store, []string{"syslog://drain-host.local"}, 3)
		rw := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

		handler.ServeHTTP(rw, req)

		resp := rw.Result()
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(http.StatusOK))

		Expect(simplifyHostnames(body)).To(MatchJSON(`{
			"results": {
				"app-id-1": {
					"drains": ["syslog://drain-host.local/?drain-version=2.0"],
					"hostname": "org.space.appname"
				},
				"app-id-2": {
					"drains": ["syslog://drain-host.local/?drain-version=2.0"],
					"hostname": "org.space.appname"
				},
				"app-id-3": {
					"drains": ["syslog://drain-host.local/?drain-version=2.0"],
					"hostname": "org.space.appname"
				}
			}
		}`))
	})
})

func simplifyHostnames(body []byte) []byte {
	re := regexp.MustCompile(`org\.space\.appname-\d+`)
	return re.ReplaceAll(body, []byte("org.space.appname"))
}

type SpyAppIDStore struct {
	getCount int
}

func (s *SpyAppIDStore) GetN(n int) []string {
	ids := make([]string, 0, n)

	for i := 0; i < n; i++ {
		ids = append(ids, fmt.Sprint("app-id-", i+1))
	}

	return ids
}
