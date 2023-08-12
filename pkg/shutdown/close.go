package shutdown

import (
	"context"
	"sync"

	"github.com/vladislav-chunikhin/lib-go/pkg/logger"
)

type Closer interface {
	Add(f func() error)
	Wait(ctx context.Context) error
	CloseAll(log logger.Logger)
}

type Close struct {
	sync.Mutex
	once  sync.Once
	funcs []func() error
	osSig SignalTrap
}

func New() *Close {
	c := &Close{
		osSig: TermSignalTrap(),
	}
	return c
}
func (c *Close) Add(f func() error) {
	c.Lock()
	defer c.Unlock()
	c.funcs = append(c.funcs, f)
}

func (c *Close) Wait(ctx context.Context) error {
	return c.osSig.Wait(ctx)
}

func (c *Close) CloseAll(log logger.Logger) {
	c.once.Do(func() {
		c.Lock()
		defer c.Unlock()

		wg := sync.WaitGroup{}
		wg.Add(len(c.funcs))

		for i := range c.funcs {
			go func(f func() error) {
				if err := f(); err != nil {
					log.Errorf("close func error: %v", err)
				}

				wg.Done()
			}(c.funcs[i])
		}

		wg.Wait()
	})
}
