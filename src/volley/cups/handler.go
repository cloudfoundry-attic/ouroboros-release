package cups

import (
	"encoding/json"
	"fmt"
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

	drains := make([]string, 0, len(h.drainURLs))
	for _, d := range h.drainURLs {
		drains = append(drains, fmt.Sprint(d, "/?drain-version=2.0"))
	}

	for _, id := range appIDs {
		bindings[id] = map[string]interface{}{
			"drains":   drains,
			"hostname": "org.space.appname",
		}
	}

	return map[string]interface{}{
		"results": bindings,
	}
}
