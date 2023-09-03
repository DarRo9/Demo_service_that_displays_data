package cache

import (
	"sync"
	"time"
)

const defaultTtl = 30 * time.Minute
const cleanInterval = time.Second

type Item struct {
	Value      interface{}
	created    time.Time
	expiration int64 //время истечения
}

type MemoryCache struct {
	sync.RWMutex
	cash map[string]Item
}

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
}

func (m *MemoryCache) Set(key string, value interface{}, ttl time.Duration) {
	m.Lock()
	defer m.Unlock()

	if ttl <= 0 {
		ttl = defaultTtl
	}

	created := time.Now()

	m.cash[key] = Item{
		value,
		created,
		created.Add(ttl).Unix(),
	}
}

func (m *MemoryCache) Get(key string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()

	item, isHas := m.cash[key]

	if isHas && time.Now().Unix() <= item.expiration {
		return item.Value, true
	}

	return nil, false
}

func (m *MemoryCache) Delete(key string) {
	m.Lock()
	defer m.Unlock()

	if _, isHas := m.cash[key]; isHas {
		delete(m.cash, key)
	}
}

//Создаёт in memory db и запускает сборщик мусора
func InitializeMemoryCache() *MemoryCache {
	cash := &MemoryCache{cash: make(map[string]Item)}
	go cash.startGC()
	return cash
}

//Каждую секунду удаляет все элементы из карты 
func (m *MemoryCache) startGC() {
	for {
		<-time.After(cleanInterval)

		if m.cash == nil {
			return
		}

		if keys := m.getExpiredKeys(); len(keys) != 0 {
			m.clearItems(keys)
		}
	}
}

//Получает информацию из карты, сохраняет ключи в список, пока не истекло время
func (m *MemoryCache) getExpiredKeys() (keys []string) {
	m.RLock()
	defer m.RUnlock()

	for key, item := range m.cash {
		if time.Now().Unix() > item.expiration {
			keys = append(keys, key)
		}
	}

	return
}

//Удаление элементов in memory db
func (m *MemoryCache) clearItems(keys []string) {
	m.Lock()
	defer m.Unlock()
	for _, k := range keys {
		delete(m.cash, k)
	}
}
