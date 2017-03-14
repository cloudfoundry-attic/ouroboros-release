package app_test

import (
	loggregator "loggregator/v2"
	"math/rand"
	"sync"
	"volley/app"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EgressV2", func() {
	It("opens specified number of firehose conns to RLP", func() {
		connectionManager := &spyConnManager{}
		store := &spyIDStore{}
		egress := app.NewEgressV2(connectionManager, store, 10, 0, 0)
		go egress.Start()

		Eventually(connectionManager.FirehoseCount).Should(Equal(10))
	})

	Context("when the filter specifes source id", func() {
		It("opens app streams using available appIDs", func() {
			availableApps := []string{"app-id-1", "app-id-2"}
			idStore := &spyIDStore{
				appIDs: availableApps,
			}
			connectionManager := &spyConnManager{}
			egress := app.NewEgressV2(connectionManager, idStore, 0, 3, 0)
			go egress.Start()

			f := func() int {
				return connectionManager.SourceCount(availableApps)
			}
			Eventually(f).Should(Equal(3))
		})
	})

	Context("when the filter is log specific", func() {
		It("opens app log streams", func() {
			availableApps := []string{"app-id-1", "app-id-2"}
			idStore := &spyIDStore{
				appIDs: availableApps,
			}
			connectionManager := &spyConnManager{}
			egress := app.NewEgressV2(connectionManager, idStore, 0, 0, 7)
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
	filters []*loggregator.Filter
	mu      sync.Mutex
}

func (s *spyConnManager) Assault(filter *loggregator.Filter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.filters = append(s.filters, filter)
}

func (s *spyConnManager) FirehoseCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	var count int
	emptyFilter := loggregator.Filter{}
	for _, f := range s.filters {
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
	for _, f := range s.filters {
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
	for _, f := range s.filters {
		for _, id := range apps {
			if f.SourceId == id && f.GetLog() != nil {
				count++
			}
		}
	}

	return count
}
