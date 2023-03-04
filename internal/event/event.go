package event

type Event struct {
	Name  string
	Value []byte
}

type Bus struct {
	C           chan Event
	subscribers map[string][]func()
}

func Init() (Bus, error) {
	eventBus := Bus{}
	eventBus.C = make(chan Event)
	eventBus.subscribers = make(map[string][]func())
	go eventBus.listen()
	return eventBus, nil
}

func (eb *Bus) listen() {
	for event := range eb.C {
		for _, emit := range eb.subscribers[event.Name] {
			go emit()
		}
	}
}

func (eb *Bus) Subscribe(eventName string, callback func()) {
	eb.subscribers[eventName] = append(eb.subscribers[eventName], callback)
}

func (eb *Bus) Publish(eventName string, data []byte) {
	eb.C <- Event{eventName, data}
}
