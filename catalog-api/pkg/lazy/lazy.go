package lazy

import (
	"sync"
)

type Value[T any] struct {
	once   sync.Once
	value  T
	err    error
	loader func() (T, error)
}

func NewValue[T any](loader func() (T, error)) *Value[T] {
	return &Value[T]{
		loader: loader,
	}
}

func (v *Value[T]) Get() (T, error) {
	v.once.Do(func() {
		v.value, v.err = v.loader()
	})
	return v.value, v.err
}

func (v *Value[T]) MustGet() T {
	val, err := v.Get()
	if err != nil {
		panic(err)
	}
	return val
}

func (v *Value[T]) Reset() {
	v.once = sync.Once{}
}

type Service[T any] struct {
	once    sync.Once
	service T
	initErr error
	init    func() (T, error)
}

func NewService[T any](init func() (T, error)) *Service[T] {
	return &Service[T]{
		init: init,
	}
}

func (s *Service[T]) Get() (T, error) {
	s.once.Do(func() {
		s.service, s.initErr = s.init()
	})
	return s.service, s.initErr
}

func (s *Service[T]) Initialized() bool {
	return s.initErr == nil
}
