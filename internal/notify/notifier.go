package notify

// Notification represents a push notification to send.
type Notification struct {
	Title    string
	Message  string
	LEDColor string
	Group    string
	ID       string
	Priority string
	Sound    bool
}

// Notifier sends notifications.
type Notifier interface {
	Send(n Notification) error
	SendBackground(n Notification)
}
