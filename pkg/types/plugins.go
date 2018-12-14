package types

// ExitSignal is a function which runs as a background process and sends a signal over a channel when an exit condition is encountered.
// If the exit path is abnormal, a non-nil error will be sent over the channel. Otherwise, for normal exit conditions, nil will be sent.
type ExitSignal func() chan error
