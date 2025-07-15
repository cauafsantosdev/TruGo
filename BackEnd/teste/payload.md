# Payloads

## Partidas

### 1. Listar Salas
```js
{
    type: "LISTAR_SALAS"
}
```

### 2. Criar Salas
```js
{
    type: "CRIAR_SALA"
    id: "ID_SALA_AQUI"
}
```

### 3. Entrar Sala
```js
{
    type: "ENTRAR_SALA",
    nome: "NOME_JOGADOR",
    idSala: "ID_SALA_AQUI",
}
```

### 4. Escolher Time
```js
{
    type: "ENTRAR_EQUIPE",
    idSala: "ID_SALA_AQUI",
    timeEscolhido: "TIME_0X" // TIME_01 ou TIME_02
}
```

## Jogadas
Esses payloads/eventos só são possiveis ser aceitos pelo servidor, se a partida/rodada começar.
O indicativo que a partida começou ou uma nova rodada iniciou, o servidor vai enviar type: "MAO_RODADA", esse payload,
que é enviado do servidor para o cliente, contem a mão do jogador (as cartas do jogador)
### 1. Fazer Jogada
```js
{
    type: "FAZER_JOGADA",
    idSala: "ID_SALA_AQUI",
    cartaJogada: {
        naipe: "NAIPE_AQUI"
        valor: "INT_DO_VALOR"
        forca: "INT_DA_FORCA"
    }
}
// SÓ É POSSÍVEL FAZER_JOGADA SE RECEBER UMA MENSAGEM COM O type: "SUA_VEZ"
```

### 2. Fazer Aposta
```js
{
    type: "TIPO_APOSTA",
    idSala: "ID_SALA_AQUI"
}

APOSTAS_POSSIVEIS = [
    "CHAMAR_TRUCO", "CHAMAR_RETRUCO", "CHAMAR_VALE_QUATRO",
    "CHAMAR_ENVIDO", "CHAMAR_REAL_ENVIDO", "CHAMAR_FALTA_ENVIDO",
    "CANTAR_FLOR",
    ]

// SÓ É POSSÍVEL FAZER APOSTA, SE RECEBER UMA MENSAGEM COM O type: "SUA_VEZ" E O TIPO DE APOSTA FOR TRUE (Todos os tipos de aposta, exceto o RETRUCO e VALE_QUATRO)
```

### 3. Aceitar Aposta
```js
{
    type: "ACEITAR_APOSTA",
    tipoAposta: "TIPO_APOSTA_AQUI",
    idSala: "ID_SALA_AQUI",
    aceitar: bool // true(quero) ou false(não quero), exceto a FLOR
}

TIPO_APOSTA = [
    "TRUCO", "RETRUCO", "VALE_QUATRO",
    "ENVIDO", "REAL_ENVIDO", "FALTA_ENVIDO",
    "FLOR",
    ]

// SÓ É POSSÍVEL ACEITAR UMA APOSTA AO RECEBER UM type: "APOSTA", ISSO INDICA QUE O JOGADOR RECEBEU UMA APOSTA, ABAIXO ESTÁ
// UM EXEMPLO DO PAYLOAD type: "APOSTA"
```
- Payload de APOSTA (quando o cliente recebe uma aposta do outro)
```js
{
    type: "APOSTA",
    tipoAposta: "TIPO_APOSTA_AQUI"
}
```

## Payloads que o servidor envia para o cliente

### 1. Vagas de equipe ao entrar na sala

```js
{
   type: "OK",
   idSala: "ID_SALA",
   Equipe01Vagas: 0 ou 1
   Equipe02Vagas: 0 ou 1 
}
// Mostra as vagas de cada equipe ao ENTRAR_SALA
```

### 2. Mão rodada

```js
{
    type: "MAO_RODADA",
    mao: [...] // Aqui vão estar as 3 cartas que o jogador recebe a cada rodada
    // seguindo o objeto abaixo
}
```
- Carta
```js
{
    naipe: "NAIPE"
    valor: number
    forca: number
}
```

### 3. SUA_VEZ e STATUS_PARTIDA
```js
    type: "SUA_VEZ" ou "STATUS_PARTIDA"
    cartasJogadas: [...] 
    apostasDisponiveis: {
        "Flor":        bool,
		"Envido":      bool,
		"Truco":       bool,
		"ContraFlor":  bool,
		"RealEnvido":  bool,
		"FaltaEnvido": bool,
		"Retruco":     bool,
		"ValeQuatro":  bool, 
    }
    placar: { // Pontos de cada equipe aqui
        TIME_01: number,
        TIME_01: number,
    }
```

- Carta Jogada
```js
    {
        idEquipe: "TIME_01" ou "TIME_02",
        jogador: "NOME_JOGADOR",
        cartaJogada: carta
    }
```

- Carta
```js
{
    naipe: "NAIPE"
    valor: number
    forca: number
}
```

### 4. Aposta e RespostaAposta

Aposta
```js
{
    type: "APOSTA"
    tipoAposta: "TIPO_APOSTA"
}
```

RespostaAposta
```js
{
    type: "RESPOSTA_APOSTA"
    tipoAposta: "TIPO_APOSTA"
    aceito: bool
}
```