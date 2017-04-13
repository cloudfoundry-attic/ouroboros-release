package cups

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

type idGetter interface {
	GetN(n int) (id []string)
}

type CUPSHandler struct {
	idGetter   idGetter
	drainURLs  []string
	drainCount int
}

func NewCUPSHandler(i idGetter, drainURLs []string, drainCount int) *CUPSHandler {
	return &CUPSHandler{
		idGetter:   i,
		drainURLs:  drainURLs,
		drainCount: drainCount,
	}
}

func (h *CUPSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := h.newResponse()

	err := json.NewEncoder(w).Encode(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *CUPSHandler) newResponse() map[string]interface{} {
	bindings := make(map[string]interface{})

	appIDs := h.idGetter.GetN(h.drainCount)

	for _, id := range appIDs {
		drain := h.drainURLs[rand.Intn(len(h.drainURLs))]

		bindings[id] = map[string]interface{}{
			"drains": []string{
				fmt.Sprint(drain, "/?drain-version=2.0"),
			},
			"hostname": "org.space.appname",
		}
	}

	return map[string]interface{}{
		"results": bindings,
	}
}
