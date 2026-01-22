package mu

import "sync"

type MutexByKey map[string]*sync.RWMutex

func (mbk MutexByKey) GetOrCreate(key string) *sync.RWMutex {
	if _, ok := mbk[key]; !ok {
		mbk[key] = &sync.RWMutex{}
	}

	return mbk[key]
}
