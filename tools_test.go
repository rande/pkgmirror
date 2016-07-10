// Copyright © 2016 Thomas Rabaix <thomas.rabaix@gmail.com>.
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pkgmirror

import (
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Compress(t *testing.T) {
	d := []byte("Hello")

	c, err := Compress(d)

	assert.NoError(t, err)
	assert.True(t, len(c) > 0)
}

func Test_Compress_EmptyData(t *testing.T) {
	d := []byte("")

	c, err := Compress(d)

	assert.NoError(t, err)
	assert.True(t, len(c) > 0)
}

func Test_WorkerManager_WorkerNumber(t *testing.T) {
	// should be called 10 times
	var cpt int32

	m := NewWorkerManager(10, func(id int, data <-chan interface{}, result chan interface{}) {
		atomic.AddInt32(&cpt, 1)
	})

	m.Start()

	m.Wait()

	assert.Equal(t, cpt, int32(10))
}

type chnStruct struct {
	v int32
}

func Test_WorkerManager_DataIn(t *testing.T) {
	// should be called 10 times
	var cpt int32

	m := NewWorkerManager(5, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			atomic.AddInt32(&cpt, raw.(chnStruct).v)
		}
	})

	m.Start()

	m.Add(chnStruct{v: 5})
	m.Add(chnStruct{v: 5})
	m.Add(chnStruct{v: 5})

	m.Wait()

	assert.Equal(t, cpt, int32(15))
}

func Test_WorkerManager_Result(t *testing.T) {
	var cpt int32

	m := NewWorkerManager(5, func(id int, data <-chan interface{}, result chan interface{}) {
		for raw := range data {
			result <- raw
		}
	})

	m.ResultCallback(func(raw interface{}) {
		cpt += raw.(chnStruct).v
	})

	m.Start()

	m.Add(chnStruct{v: 5})
	m.Add(chnStruct{v: 5})
	m.Add(chnStruct{v: 5})

	m.Wait()

	assert.Equal(t, int32(15), cpt)
}
