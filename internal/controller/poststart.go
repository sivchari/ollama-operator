package controller

var postStartScript = "{{- range .Images }}ollama pull {{ . }};{{- end }}"

type PostStartInput struct {
	Images []string
}
