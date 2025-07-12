package models

import "fmt"

type Cartas struct {
	Valor int
	Naipe string
	Forca int
}

type CartaJogada struct {
	Jogador *Jogador
	Carta   *Cartas
}

func (c Cartas) String() string {
	return fmt.Sprintf("%d de %s", c.Valor, c.Naipe)
}
