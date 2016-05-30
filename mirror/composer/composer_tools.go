// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package composer

import "sync"

type DownloadManager struct {
	Add   chan PackageInformation
	Count int
	Done  chan struct{}
}

func (dm *DownloadManager) Wait(fn func(id int, urls <-chan PackageInformation)) {
	var wg sync.WaitGroup

	wg.Add(dm.Count)

	pkgs := make(chan PackageInformation)

	for i := 0; i < dm.Count; i++ {
		go func(id int) {
			fn(id, pkgs)
			wg.Done()
		}(i)
	}

	go func() {
		// receive url
		for {
			select {
			case pkg := <-dm.Add:
				pkgs <- pkg

			case <-dm.Done:
				close(pkgs)
				return
			}
		}
	}()

	wg.Wait()
}