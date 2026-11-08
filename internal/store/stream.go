package store

type streamer[T any] struct {
	publish chan T
	sub     chan chan T
	unsub   chan chan T
	cancel  chan struct{}
}

func Streamer[T any]() *streamer[T] {
	return &streamer[T]{
		publish: make(chan T, 1),
		sub:     make(chan chan T, 1),
		unsub:   make(chan chan T, 1),
	}
}

func (s *streamer[T]) Start() {
	subs := map[chan T]struct{}{}

	go func() {
		for {
			select {
			case <-s.cancel:
				return
			case ch := <-s.sub:
				subs[ch] = struct{}{}
			case ch := <-s.unsub:
				delete(subs, ch)
			case msg := <-s.publish:
				for ch := range subs {
					select {
					case ch <- msg:
					default:
					}
				}
			}
		}
	}()

}

func (s *streamer[T]) Stop() {
	close(s.cancel)
}

func (s *streamer[T]) Subscribe() chan T {
	ch := make(chan T, 5)
	s.sub <- ch
	return ch
}

func (s *streamer[T]) Unsubscribe(ch chan T) {
	s.unsub <- ch
}

func (s *streamer[T]) Publish(msg T) {
	s.publish <- msg
}
