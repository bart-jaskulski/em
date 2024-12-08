package types

type EmojiLoadedMsg struct {
    Emojis map[string][]string
    Err    error
}

type State int

const (
    StateLoading State = iota
    StateReady
    StateError
)