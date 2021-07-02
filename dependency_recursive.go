package king

import (
	"errors"
	"fmt"
	"os"
)

const (
	white = iota
	black
	grey
)

type graph map[string]*vertex

type vertex struct {
	n string
	d *Dependency

	cl int
	tg graph
	ee map[string]*vertex
}

func (tg graph) addVertex(n string) *vertex {
	if vx, ok := tg[n]; ok {
		return vx
	}

	vx := &vertex{
		n:  n,
		tg: tg,
		ee: make(map[string]*vertex),
	}

	tg[n] = vx
	return vx
}

func (vx *vertex) addEdge(d *Dependency) bool {
	if eg, ok := vx.ee[d.Name]; ok {
		if eg.d != nil && eg.d.IsMake && !d.IsMake {
			eg.d.IsMake = false
		}

		return false
	}

	eg := vx.tg.addVertex(d.Name)
	eg.d = d // TODO

	vx.ee[d.Name] = eg
	return true
}

func (tg graph) populate(p *Package) error {
	dd, err := p.Dependencies()

	if err != nil {
		return err
	}

	vx := tg.addVertex(p.Name)

	for _, d := range dd {
		if !vx.addEdge(d) {
			continue
		}

		dp, err := NewPackage(p.cfg, &PackageOptions{
			Name: d.Name,
			From: p.From,
		})

		if err != nil {
			return err
		}

		if err := tg.populate(dp); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return err
		}
	}

	return nil
}

func (vx *vertex) traversal(dd *[]*Dependency) error {
	if vx.cl == black {
		return nil
	}

	vx.cl = grey

	for _, eg := range vx.ee {
		switch eg.cl {
		case white:
			if err := eg.traversal(dd); err != nil {
				return err
			}
		case black:
			continue
		case grey:
			return fmt.Errorf("cycle between %s and %s", vx.n, eg.n)
		}
	}

	if vx.d != nil {
		*dd = append(*dd, vx.d)
	}

	vx.cl = black
	return nil
}

func (p *Package) RecursiveDependencies() ([]*Dependency, error) {
	tg := make(graph)

	if err := tg.populate(p); err != nil {
		return nil, err
	}

	dd := make([]*Dependency, 0, len(tg))

	for _, vx := range tg {
		if err := vx.traversal(&dd); err != nil {
			panic(fmt.Sprintf("parse %s reverse dependencies: %v", p.Name, err))
		}
	}

	return dd, nil
}
