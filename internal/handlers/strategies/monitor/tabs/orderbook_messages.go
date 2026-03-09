package tabs

// backToExchangeListMsg signals to go back to the exchange list view.
// source identifies which tab coordinator owns this message.
type backToExchangeListMsg struct {
	source string
}
