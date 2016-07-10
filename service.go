// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"sync"

	"github.com/rande/goapp"
)

const (
	STATUS_RUNNING = 1
	STATUS_HOLD    = 2
	STATUS_ERROR   = 3
)

type MirrorService interface {
	Init(app *goapp.App) error
	Serve(state *goapp.GoroutineState) error
}

type ServiceRegistry struct {
	Services map[string]MirrorService
	lock     *sync.Mutex
}

func (sr *ServiceRegistry) Add(name string, s MirrorService) {
	sr.lock.Lock()
	sr.Services[name] = s
	sr.lock.Unlock()
}

type State struct {
	Id      string
	Status  int
	Message string
}
