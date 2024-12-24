package optimization

import (
	"sync"

	"github.com/notnil/chess"
)

// GameStatePool пул для переиспользования игровых состояний
type GameStatePool struct {
	pool sync.Pool
}

// NewGameStatePool создает новый пул состояний
func NewGameStatePool() *GameStatePool {
	return &GameStatePool{
		pool: sync.Pool{
			New: func() interface{} {
				return &chess.Game{}
			},
		},
	}
}

// Get получает состояние из пула
func (p *GameStatePool) Get() *chess.Game {
	return p.pool.Get().(*chess.Game)
}

// Put возвращает состояние в пул
func (p *GameStatePool) Put(game *chess.Game) {
	p.pool.Put(game)
}

// ObjectPool общий пул объектов
type ObjectPool struct {
	pool sync.Pool
	new  func() interface{}
}

// NewObjectPool создает новый пул объектов
func NewObjectPool(new func() interface{}) *ObjectPool {
	return &ObjectPool{
		pool: sync.Pool{New: new},
	}
}

// Get получает объект из пула
func (p *ObjectPool) Get() interface{} {
	return p.pool.Get()
}

// Put возвращает объект в пул
func (p *ObjectPool) Put(obj interface{}) {
	p.pool.Put(obj)
}
