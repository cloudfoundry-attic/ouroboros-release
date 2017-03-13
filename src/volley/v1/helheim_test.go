// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package v1_test

import "github.com/cloudfoundry/dropsonde/metricbatcher"

type mockAppIDStore struct {
	AddCalled chan bool
	AddInput  struct {
		AppID chan string
	}
	GetCalled chan bool
	GetOutput struct {
		Ret0 chan string
	}
}

func newMockAppIDStore() *mockAppIDStore {
	m := &mockAppIDStore{}
	m.AddCalled = make(chan bool, 100)
	m.AddInput.AppID = make(chan string, 100)
	m.GetCalled = make(chan bool, 100)
	m.GetOutput.Ret0 = make(chan string, 100)
	return m
}
func (m *mockAppIDStore) Add(appID string) {
	m.AddCalled <- true
	m.AddInput.AppID <- appID
}
func (m *mockAppIDStore) Get() string {
	m.GetCalled <- true
	return <-m.GetOutput.Ret0
}

type mockBatcher struct {
	BatchCounterCalled chan bool
	BatchCounterInput  struct {
		Name chan string
	}
	BatchCounterOutput struct {
		Ret0 chan metricbatcher.BatchCounterChainer
	}
}

func newMockBatcher() *mockBatcher {
	m := &mockBatcher{}
	m.BatchCounterCalled = make(chan bool, 100)
	m.BatchCounterInput.Name = make(chan string, 100)
	m.BatchCounterOutput.Ret0 = make(chan metricbatcher.BatchCounterChainer, 100)
	return m
}
func (m *mockBatcher) BatchCounter(name string) metricbatcher.BatchCounterChainer {
	m.BatchCounterCalled <- true
	m.BatchCounterInput.Name <- name
	return <-m.BatchCounterOutput.Ret0
}

type mockBatchCounterChainer struct {
	SetTagCalled chan bool
	SetTagInput  struct {
		Key, Value chan string
	}
	SetTagOutput struct {
		Ret0 chan metricbatcher.BatchCounterChainer
	}
	IncrementCalled chan bool
	AddCalled       chan bool
	AddInput        struct {
		Value chan uint64
	}
}

func newMockBatchCounterChainer() *mockBatchCounterChainer {
	m := &mockBatchCounterChainer{}
	m.SetTagCalled = make(chan bool, 100)
	m.SetTagInput.Key = make(chan string, 100)
	m.SetTagInput.Value = make(chan string, 100)
	m.SetTagOutput.Ret0 = make(chan metricbatcher.BatchCounterChainer, 100)
	m.IncrementCalled = make(chan bool, 100)
	m.AddCalled = make(chan bool, 100)
	m.AddInput.Value = make(chan uint64, 100)
	return m
}
func (m *mockBatchCounterChainer) SetTag(key, value string) metricbatcher.BatchCounterChainer {
	m.SetTagCalled <- true
	m.SetTagInput.Key <- key
	m.SetTagInput.Value <- value
	return <-m.SetTagOutput.Ret0
}
func (m *mockBatchCounterChainer) Increment() {
	m.IncrementCalled <- true
}
func (m *mockBatchCounterChainer) Add(value uint64) {
	m.AddCalled <- true
	m.AddInput.Value <- value
}