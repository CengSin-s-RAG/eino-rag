package runtime

import "errors"

// Closer is deliberately small: providers, stores and external clients can all
// participate in application shutdown without knowing about each other.
type Closer interface {
	Close() error
}

type App struct {
	closers []Closer
}

func NewApp(closers ...Closer) *App {
	return &App{closers: closers}
}

func (a *App) Close() error {
	var errs []error
	for i := len(a.closers) - 1; i >= 0; i-- {
		if a.closers[i] != nil {
			errs = append(errs, a.closers[i].Close())
		}
	}
	return errors.Join(errs...)
}
