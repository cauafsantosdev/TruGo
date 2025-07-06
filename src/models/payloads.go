package models

type Payload struct {
	Type string `json:"type"`
}
type Resposta struct {
	Type string `json:"type"`
	Msg string `json:"message"`
}

type EntrarSala struct {
	Nome   string `json:"nome"`
	IdSala string `json:"idSala"`
}

type CriarSalaID struct {
	ID string `json:"id"`
}