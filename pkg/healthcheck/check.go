package healthcheck

import "context"

func ContextDone(ctx context.Context) func() error {
	return func() error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			return nil
		}
	}
}
