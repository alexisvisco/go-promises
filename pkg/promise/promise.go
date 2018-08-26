package promise

import (
	"sync"
)

type Resolve func(interface{}) error
type Reject func(error) error
type Fulfill func(resolve Resolve, reject Reject)

type Promiseable interface {
	Then(resolve Resolve) Promiseable
	Catch(reject Reject) Promiseable
}

type Resolver struct {
	resolve Resolve
	*Promise
}

type Rejecter struct {
	reject Reject
	*Promise
}

type Promise struct {
	resolved []*Resolver
	rejected []*Rejecter

	*sync.Mutex
}

func New(fulfill Fulfill) *Promise {
	p := newPromise()
	go fulfill(p.resolve, p.reject)
	return p
}

func newPromise() *Promise {
	p := &Promise{
		resolved: make([]*Resolver, 0),
		rejected: make([]*Rejecter, 0),
		Mutex:    &sync.Mutex{},
	}
	return p
}

func (p *Promise) resolve(i interface{}) error {
	p.Lock()
	defer p.Unlock()
	for _, res := range p.resolved {
		if err := res.resolve(i); err != nil {
			res.reject(err)
		} else {
			res.resolve(nil)
		}
	}
	return nil
}

func (p *Promise) reject(err error) error {
	p.Lock()
	defer p.Unlock()
	for _, rej := range p.rejected {
		if err := rej.reject(err) ; err != nil{
			rej.reject(err)
		} else {
			rej.resolve(nil)
		}
	}
	return nil
}

func (p *Promise) Then(resolve Resolve) Promiseable {
	p.Lock()
	defer p.Unlock()
	resolver := &Resolver{
		resolve: resolve,
		Promise: newPromise(),
	}
	p.resolved = append(p.resolved, resolver)
	return resolver
}

func (p *Promise) Catch(reject Reject) Promiseable {
	p.Lock()
	defer p.Unlock()
	catcher := &Rejecter{
		reject: reject,
		Promise: newPromise(),
	}
	p.rejected = append(p.rejected, catcher)
	return p
}