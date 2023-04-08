package event

type Event struct {
	Name string
	Data []byte
}

var B *Bus

type subscriber func([]byte)

type Bus struct {
	c           chan Event
	subscribers map[string][]subscriber
}

func Init() error {
	B = &Bus{
		c:           make(chan Event),
		subscribers: make(map[string][]subscriber),
	}
	go B.listen()
	return nil
}

func (eb *Bus) listen() {
	for event := range eb.c {
		for _, emit := range eb.subscribers[event.Name] {
			go emit(event.Data)
		}
	}
}

func (eb *Bus) Subscribe(eventName string, callback func([]byte)) {
	eb.subscribers[eventName] = append(eb.subscribers[eventName], callback)
}

func (eb *Bus) Publish(eventName string, data []byte) {
	eb.c <- Event{eventName, data}
}
