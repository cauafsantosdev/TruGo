package models

type Payload struct {
	Type string `json:"type"`
}

type Resposta struct {
	Type string `json:"type"`
	Msg  string `json:"message"`
}

type EntrarSala struct {
	Nome   string `json:"nome"`
	IdSala string `json:"idSala"`
}

type EscolherEquipe struct {
	ID            string `json:"idSala"`
	TimeEscolhido string `json:"timeEscolhido"`
}

type CriarSalaID struct {
	ID string `json:"id"`
}

type EntrouSalaResposta struct {
	Type          string `json:"type"`
	ID            string `json:"idSala"`
	Equipe01Vagas int    `json:"Equipe01Vagas"`
	Equipe02Vagas int    `json:"Equipe02Vagas"`
}

type SalasDisponiveis struct {
	SalasDisponiveis map[string]int `json:"salasDisponiveis"`
}

type CartaResposta struct {
	Valor int    `json:"valor"`
	Naipe string `json:"naipe"`
	Forca int    `json:"forca"`
}

type Jogada struct {
	IDEquipe    string `json:"idEquipe"`
	JogadorNome string `json:"jogador"`
	CartaJogada CartaResposta `json:"cartaJogada"`
}

type StatusRodada struct {
	Type              string          `json:"type"`
	CartasJogadas     []Jogada        `json:"cartasJogadas"`
	ApostasDiponiveis map[string]bool `json:"apostasDisponiveis"`
}
