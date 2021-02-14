package once

import (
	"sync/atomic"
)

type Once struct {
	c uint32
}

func (o *Once) Do(f func() error) error {
	if atomic.LoadUint32(&o.c) == 1 {
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	atomic.StoreUint32(&o.c, 1)
	return nil
}

func (o *Once) Reset() {
	atomic.StoreUint32(&o.c, 0)
}
