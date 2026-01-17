package mu

import "sync"

type FileMutexByKey map[string]*sync.RWMutex

func (fmkb FileMutexByKey) GetOrCreate(key string) *sync.RWMutex {
	if _, ok := fmkb[key]; !ok {
		fmkb[key] = &sync.RWMutex{}
	}

	return fmkb[key]
}
