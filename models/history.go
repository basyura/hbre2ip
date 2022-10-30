package models

type History struct {
	Entries []Entry `json:"entries"`
}

func (h *History) Contains(entry Entry) bool {
	for _, e := range h.Entries {
		if e.Url == entry.Url {
			return true
		}
	}
	return false
}

func (h *History) Add(entry Entry) {
	h.Entries = append(h.Entries, entry)
}
