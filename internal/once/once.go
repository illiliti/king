package once

type Once bool

func (o *Once) Do(f func() error) error {
	if *o {
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	*o = true
	return nil
}

func (o *Once) Reset() {
	*o = false
}
