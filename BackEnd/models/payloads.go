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

type PartidaFinalizada struct {
	Type     string         `json:"type"`
	Mensagem string         `json:"message"`
	Placar   map[string]int `json:"placar"`
}

type CartaResposta struct {
	Valor int    `json:"valor"`
	Naipe string `json:"naipe"`
	Forca int    `json:"forca"`
}

type Jogada struct {
	IDEquipe    string        `json:"idEquipe"`
	JogadorNome string        `json:"jogador"`
	CartaJogada CartaResposta `json:"cartaJogada"`
}

type StatusRodada struct {
	Type              string          `json:"type"`
	CartasJogadas     []Jogada        `json:"cartasJogadas"`
	ApostasDiponiveis map[string]bool `json:"apostasDisponiveis"`
	Placar            map[string]int  `json:"placar"`
}

type FazerJogada struct {
	Type         string        `json:"type"`
	IDSala       string        `json:"idSala"`
	CartaJogada  CartaResposta `json:"cartaJogada"`
	ApostaPedida string        `json:"apostaPedida"`
}

type IDSala struct {
	Type   string `json:"type"`
	IDSala string `json:"idSala"`
}

type EnviarAposta struct {
	Type         string `json:"type"`
	TipoDeAposta string `json:"aposta"`
}

type ResponderAposta struct {
	Type       string `json:"type"`
	TipoAposta string `json:"tipoAposta"`
	IDSala     string `json:"idSala"`
	Aceitar    bool   `json:"aceitar"`
}

type RespostaAposta struct {
	Type       string `json:"type"`
	TipoAposta string `json:"tipoAposta"`
	Quero      bool   `json:"quero"`
}

type MaoDaRodada struct {
	Type    string            `json:"type"`
	Mao     []CartaResposta   `json:"mao"`
	Equipes map[string]string `json:"equipes"`
}

type MaoFinalizada struct {
	Type          string         `json:"type"`
	TimeVencedor  string         `json:"timeVencedor"`
	PontosGanhos  int            `json:"pontosGanhos"`
	Placar        map[string]int `json:"placar"`
	CartasJogadas []Jogada       `json:"cartasJogadas"`
}

type ComandoAudio struct {
	Type      string `json:"type"`
	IDSala    string `json:"idSala"`
	NomeAudio string `json:"nomeAudio"`
}

type TocarAudio struct {
	Type      string `json:"type"`
	NomeAudio string `json:"nomeAudio"`
}

type RespostaFlor struct {
	Type             string `json:"type"`
	RespostaParaFlor bool   `json:"boa"`
}

type RespostaFlorAdversario struct {
	Type             string `json:"type"`
	RespostaParaFlor bool   `json:"apostaFlor"`
}

type PontosDaMao struct {
	Type   string `json:"type"`
	Vencedor string `json:"vencedor"`
	Equipe map[string]int
	Placar map[string]int `json:"placar"`
}
