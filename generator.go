package genny

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
)

// Generator is the basic type for generators to use
type Generator struct {
	StepName     string
	Should       func(*Runner) bool
	Root         string
	runners      []RunFn
	transformers []Transformer
	moot         *sync.RWMutex
}

// New, well-formed, generator
func New() *Generator {
	g := &Generator{
		StepName:     fmt.Sprint(time.Now().UnixNano()),
		runners:      []RunFn{},
		moot:         &sync.RWMutex{},
		transformers: []Transformer{},
	}
	return g
}

// File adds a file to be run when the generator is run
func (g *Generator) File(f File) {
	g.RunFn(func(r *Runner) error {
		return r.File(f)
	})
}

func (g *Generator) Transform(f File) (File, error) {
	g.moot.RLock()
	defer g.moot.RUnlock()
	var err error
	for _, t := range g.transformers {
		f, err = t.Transform(f)
		if err != nil {
			return f, errors.WithStack(err)
		}
	}

	return f, nil
}

// Transformer adds a file transform to the generator
func (g *Generator) Transformer(t Transformer) {
	g.moot.Lock()
	defer g.moot.Unlock()
	g.transformers = append(g.transformers, t)
}

// Command adds a command to be run when the generator is run
func (g *Generator) Command(cmd *exec.Cmd) {
	g.RunFn(func(r *Runner) error {
		return r.Exec(cmd)
	})
}

// Box walks through a packr.Box and adds Files for each entry
// in the box.
func (g *Generator) Box(box packr.Box) error {
	return box.Walk(func(path string, f packr.File) error {
		g.File(NewFile(path, f))
		return nil
	})
}

// RunFn adds a generic "runner" function to the generator.
func (g *Generator) RunFn(fn RunFn) {
	g.moot.Lock()
	defer g.moot.Unlock()
	g.runners = append(g.runners, fn)
}

func (g1 *Generator) Merge(g2 *Generator) {
	g2.moot.Lock()
	g1.moot.Lock()
	g1.runners = append(g1.runners, g2.runners...)
	g1.transformers = append(g1.transformers, g2.transformers...)
	g1.moot.Unlock()
	g2.moot.Unlock()
}
