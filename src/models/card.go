package models

type Cartas struct {
	Valor int
	Naipe string
	Forca int
}

type CartaJogada struct {
	Jogador *Jogador
	Carta   *Cartas
}
