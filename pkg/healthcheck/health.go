package healthcheck

import (
	"sync"
)

type HealthChecker interface {
	RegisterLive(name string, checker CheckerFunc)
	RegisterReady(name string, checker CheckerFunc)
	CheckLive() map[string]error
	CheckReady() map[string]error
}

type CheckerFunc func() error

// HealthCheck is a structure that performs system checks
type HealthCheck struct {
	liveCheck  map[string]CheckerFunc
	readyCheck map[string]CheckerFunc
	mux        sync.Mutex
}

// NewHealthCheck creates a new instance of HealthCheck
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		liveCheck:  make(map[string]CheckerFunc),
		readyCheck: make(map[string]CheckerFunc),
		mux:        sync.Mutex{},
	}
}

// RegisterLive adds a component check
func (i *HealthCheck) RegisterLive(name string, checker CheckerFunc) {
	if i == nil {
		return
	}
	i.mux.Lock()
	defer i.mux.Unlock()
	i.liveCheck[name] = checker
}

// RegisterReady adds a component check
func (i *HealthCheck) RegisterReady(name string, checker CheckerFunc) {
	if i == nil {
		return
	}
	i.mux.Lock()
	defer i.mux.Unlock()
	i.readyCheck[name] = checker
}

// CheckLive calls registered checks and returns the result
func (i *HealthCheck) CheckLive() map[string]error {
	if i == nil {
		return nil
	}
	return i.check(i.liveCheck)
}

// CheckReady calls registered checks and returns the result
func (i *HealthCheck) CheckReady() map[string]error {
	if i == nil {
		return nil
	}
	return i.check(i.readyCheck)
}

type result struct {
	Name string
	Err  error
}

func (i *HealthCheck) check(checks map[string]CheckerFunc) map[string]error {
	checkRes := make(chan result)

	go func() {
		wg := sync.WaitGroup{}
		defer close(checkRes)
		for name, checker := range checks {
			wg.Add(1)
			go func(name string, checker CheckerFunc) {
				defer wg.Done()
				checkRes <- result{Name: name, Err: checker()}
			}(name, checker)
		}
		wg.Wait()
	}()

	res := make(map[string]error, len(checks))
	for r := range checkRes {
		res[r.Name] = r.Err
	}

	return res
}
