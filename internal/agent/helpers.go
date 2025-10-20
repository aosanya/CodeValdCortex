package agent

// CloseDone safely closes the done channel
func (a *Agent) CloseDone() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.doneClosed {
		close(a.done)
		a.doneClosed = true
	}
}
