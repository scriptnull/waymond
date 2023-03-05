package event

type Event struct {
	Name  string
	Value []byte
}

var B *Bus

type Bus struct {
	c           chan Event
	subscribers map[string][]func()
}

func Init() error {
	B = &Bus{
		c:           make(chan Event),
		subscribers: make(map[string][]func()),
	}
	go B.listen()
	return nil
}

func (eb *Bus) listen() {
	for event := range eb.c {
		for _, emit := range eb.subscribers[event.Name] {
			go emit()
		}
	}
}

func (eb *Bus) Subscribe(eventName string, callback func()) {
	eb.subscribers[eventName] = append(eb.subscribers[eventName], callback)
}

func (eb *Bus) Publish(eventName string, data []byte) {
	eb.c <- Event{eventName, data}
}
