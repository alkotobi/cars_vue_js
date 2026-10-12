package wa

// Stub is a placeholder driver used until Stage 5 lands the rod-based browser
// layer. Every browser-backed method returns ErrNotImplemented, so the CLIs and
// MCP server still answer — with a clear error — instead of crashing. The
// offline paths (whatsapp_identity, `pending --state`) work because they never
// call the driver.
type Stub struct{}

func (Stub) Init() error             { return ErrNotImplemented }
func (Stub) RestoreAllFilter() error { return ErrNotImplemented }
func (Stub) IsLoggedIn() (bool, error) {
	return false, ErrNotImplemented
}
func (Stub) GetChats(int) ([]ChatInfo, error) { return nil, ErrNotImplemented }
func (Stub) GetLastChats(int, int) (ChatSweepResult, error) {
	return ChatSweepResult{}, ErrNotImplemented
}
func (Stub) GetLastUnreadChats(int, int) (ChatSweepResult, error) {
	return ChatSweepResult{}, ErrNotImplemented
}
func (Stub) OpenChat(string) (bool, error)    { return false, ErrNotImplemented }
func (Stub) SendMessage(string) (bool, error) { return false, ErrNotImplemented }
func (Stub) GetLastMessages(int) ([]Message, error) {
	return nil, ErrNotImplemented
}
func (Stub) MarkAsRead() (bool, error)         { return false, ErrNotImplemented }
func (Stub) SessionInfo() (SessionInfo, error) { return SessionInfo{}, ErrNotImplemented }
func (Stub) IsAttached() bool                  { return false }
func (Stub) Close() error                      { return nil }
