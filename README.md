# TruGo

Com certeza\! Aqui está a documentação formatada em Markdown.

-----

# **Documentação: Motor do Jogo de Truco (`truco.go`)**

### **Visão Geral**

Este arquivo, `truco.go`, implementa o **motor** e todas as regras centrais para uma partida de Truco Gaudério (ou Truco Argentino) para dois jogadores. Ele é projetado para ser o *backend* do jogo, gerenciando o estado, as regras, as jogadas e a pontuação de forma autônoma.

O motor é responsável por:

  * Criar e gerenciar o baralho e as cartas.
  * Controlar o estado da partida (ex: aguardando jogada, aguardando resposta de aposta, etc.).
  * Adicionar e gerenciar os jogadores e seus times.
  * Distribuir as cartas a cada mão (`mão`).
  * Processar todas as ações dos jogadores (jogar carta, cantar Truco, Envido, Flor, etc.).
  * Validar as jogadas de acordo com as regras e o turno atual.
  * Calcular e atualizar o placar.
  * Determinar os vencedores de cada rodada, mão e da partida.
  * Garantir que as operações no estado do jogo sejam seguras contra condições de corrida (*race conditions*) através de um `Mutex`.

### **Estruturas Principais (Core Structs)**

A lógica do jogo é construída em torno de algumas estruturas de dados essenciais:

  * **`Carta`**: Representa uma carta do baralho. Contém seu `Valor` (1-7, 10-12), `Naipe`, e, mais importante, sua `Forca` para as disputas de rodada. As "manilhas" (Espadão, Bastão, 7 de Espadas, 7 de Ouros) possuem os maiores valores de `Forca`.

  * **`Jogador`**: Representa um participante. Contém seu `ID`, `Nome`, `Time`, a `Mao` (as 3 cartas que ele possui) e os pontos calculados para `Envido` e `Flor`.

  * **`Acao`**: É a estrutura usada para comunicar uma jogada ao motor. Funciona como um "comando".

      * `Tipo`: Uma string que define a jogada (ex: `"JOGAR_CARTA"`, `"TRUCO"`, `"ENVIDO"`, `"QUERO"`).
      * `JogadorID`: Identifica quem está realizando a ação.
      * `Valor`: Contém dados adicionais para a ação, como o objeto `Carta` que está sendo jogado ou a string `"QUERO"`/`"NAO_QUERO"` como resposta.

  * **`Jogo`**: É a estrutura central que encapsula **todo o estado da partida**. Contém o baralho, a lista de jogadores, o placar, o turno atual, o estado da mão (quem venceu cada rodada), as apostas ativas (`Envido`, `Truco`), etc.

### **Fluxo de Jogo (Game Lifecycle)**

A interação com o motor segue um fluxo bem definido:

1.  **Criação do Jogo**:

      * Tudo começa com a chamada de `CriarJogo()`. Isso inicializa uma nova instância da struct `Jogo` no estado `AguardandoJogadores`.

2.  **Adição de Jogadores**:

      * Os jogadores são adicionados com a função `jogo.AdicionarJogador(jogador)`.
      * O motor automaticamente atribui os times (Time 1, Time 2).
      * Assim que o segundo jogador é adicionado, o jogo começa automaticamente, chamando a função `proximaMao()` pela primeira vez.

3.  **Processamento de Ações**:

      * A partir daqui, toda a interação dos jogadores com o jogo deve ocorrer através da função `jogo.ProcessarAcao(acao Acao)`.
      * Esta é a **função principal** do motor. Ela recebe um objeto `Acao`, identifica o tipo de jogada e a encaminha para a função interna correspondente (ex: `jogarCarta`, `cantarTruco`, `responder`, etc.).

4.  **Gerenciamento de Estado**:

      * O campo `jogo.Estado` (do tipo `EstadoDoJogo`) dita quais ações são permitidas em cada momento. Por exemplo, se o estado for `AguardandoRespostaTruco`, a única ação válida para o jogador do turno será uma resposta como `"QUERO"`, `"NAO_QUERO"` ou uma aposta maior (`"RETRUCO"`).
      * Após cada ação, o estado do jogo é atualizado, o turno pode passar para o próximo jogador, e o placar pode ser modificado.

5.  **Fim da Mão e do Jogo**:

      * Quando uma mão termina (seja porque um jogador correu para o baralho ou porque as três rodadas foram concluídas), o estado muda para `MaoFinalizada`.
      * O motor então verifica o placar. Se nenhum time atingiu os 30 pontos, ele automaticamente chama `proximaMao()` para preparar a rodada seguinte.
      * Se um time atingir 30 pontos, o estado muda para `FimDeJogo`.

### **Como Interagir com o Motor**

Para integrar este motor a outra parte da plataforma (como uma API ou uma interface de usuário), o fluxo de trabalho será:

1.  **Instanciar o jogo**:

    ```go
    jogo := CriarJogo()
    ```

2.  **Adicionar os jogadores**:

    ```go
    // Supondo que você tenha os objetos jogador1 e jogador2
    jogo.AdicionarJogador(jogador1)
    jogo.AdicionarJogador(jogador2)
    ```

3.  **Enviar ações e observar o estado**: O sistema externo deve construir e enviar objetos `Acao` para o motor. Após cada ação, ele deve ler o estado público do objeto `jogo` para atualizar a UI.

    **Exemplo**: O Jogador 1 (ID "j1") quer jogar a carta "3 de Espadas".

    ```go
    // A interface do usuário cria este objeto de ação
    cartaParaJogar := map[string]any{"Valor": 3, "Naipe": "Espadas"}
    acao := Acao{
        Tipo:      "JOGAR_CARTA",
        JogadorID: "j1",
        Valor:     cartaParaJogar,
    }

    // Envia a ação para o motor
    err := jogo.ProcessarAcao(acao)
    if err != nil {
        // Tratar erro (ex: jogada inválida)
    }

    // Após a ação, a interface pode ler o estado atualizado do jogo:
    // - jogo.Placar
    // - jogo.TurnoDoJogador
    // - jogo.CartasNaMesa
    // etc.
    ```

### **Regras e Lógicas Específicas Implementadas**

  * **Cálculo de Envido e Flor**: As funções `calcularPontosEnvido` e `calcularPontosFlor` são chamadas automaticamente quando as cartas são distribuídas.
  * **Hierarquia de Apostas**: O motor gerencia as cadeias de apostas para o Envido (`CadeiaDeEnvido`) e para o Truco, garantindo que apenas apostas válidas e crescentes possam ser feitas.
  * **Resolução de Rodadas e Mãos**: A função `resolverRodada` compara a `Forca` das cartas na mesa para determinar o vencedor da rodada. `resolverMao` analisa os resultados das três rodadas para encontrar o vencedor da mão, incluindo as regras complexas de empate.
  * **Regras Especiais**: O motor já implementa lógicas como "cantar com 29 pontos" (que faz o placar do time voltar para 15) e a regra opcional da "Família Real".
  * **Concorrência**: O uso de `jogo.mutex` em todas as funções que modificam o estado do jogo garante que, mesmo que as ações cheguem de forma concorrente (ex: via requisições web simultâneas), elas serão processadas uma de cada vez, evitando corrupção de dados.