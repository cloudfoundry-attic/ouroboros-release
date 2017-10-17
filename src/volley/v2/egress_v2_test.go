package v2_test

import (
	"math/rand"
	"sync"

	"volley/v2"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EgressV2", func() {
	It("opens specified number of firehose conns to RLP", func() {
		connectionManager := &spyConnManager{}
		store := &spyIDStore{}
		egress := v2.NewEgressV2(connectionManager, store, 10, 0, 0)
		go egress.Start()

		Eventually(connectionManager.FirehoseCount).Should(Equal(10))
	})

	Context("when the selectors specifes source id", func() {
		It("opens app streams using available appIDs", func() {
			availableApps := []string{"app-id-1", "app-id-2"}
			idStore := &spyIDStore{
				appIDs: availableApps,
			}
			connectionManager := &spyConnManager{}
			egress := v2.NewEgressV2(connectionManager, idStore, 0, 3, 0)
			go egress.Start()

			f := func() int {
				return connectionManager.SourceCount(availableApps)
			}
			Eventually(f).Should(Equal(3))
		})
	})

	Context("when the selector is log specific", func() {
		It("opens app log streams", func() {
			availableApps := []string{"app-id-1", "app-id-2"}
			idStore := &spyIDStore{
				appIDs: availableApps,
			}
			connectionManager := &spyConnManager{}
			egress := v2.NewEgressV2(connectionManager, idStore, 0, 0, 7)
			go egress.Start()

			f := func() int {
				return connectionManager.AppLogCount(availableApps)
			}
			Eventually(f).Should(Equal(7))
		})
	})
})

type spyIDStore struct {
	appIDs []string
}

func (s *spyIDStore) Get() string {
	if len(s.appIDs) == 0 {
		return ""
	}
	return s.appIDs[rand.Intn(len(s.appIDs))]
}

type spyConnManager struct {
	selectors []*loggregator_v2.Selector
	mu        sync.Mutex
}

func (s *spyConnManager) Assault(selector *loggregator_v2.Selector) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.selectors = append(s.selectors, selector)
}

func (s *spyConnManager) FirehoseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	emptyFilter := loggregator_v2.Selector{}
	for _, f := range s.selectors {
		if *f == emptyFilter {
			count++
		}
	}

	return count
}

func (s *spyConnManager) SourceCount(apps []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	for _, f := range s.selectors {
		for _, id := range apps {
			if f.SourceId == id && f.GetLog() == nil {
				count++
			}
		}
	}

	return count
}

func (s *spyConnManager) AppLogCount(apps []string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	for _, f := range s.selectors {
		for _, id := range apps {
			if f.SourceId == id && f.GetLog() != nil {
				count++
			}
		}
	}

	return count
}
