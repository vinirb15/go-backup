# Transcriber — Documentação de Produto

> **Para que serve este documento:** é a fonte de verdade de **negócio** do Transcriber.
> Serve de contexto para a IA gerar histórias de usuário (`/story`) e executar tarefas
> (`/work`) de forma coerente com o que o produto realmente faz. Descreve o que o usuário
> **vê e faz** na interface web — sem detalhes de implementação.

---

## 1. Visão Geral do Produto

O **Transcriber** é uma plataforma web que transcreve áudio de reuniões e conversas em texto,
em tempo real, e enriquece esse texto com correção e resumo automáticos por IA.

Todo o processamento acontece **localmente** (sem envio para a nuvem), o que torna o produto
adequado para conteúdos sensíveis e privados.

**Proposta de valor (jornada central):**

> Gravar a reunião → transcrever ao vivo → corrigir o texto → gerar um resumo estruturado.

**Diferenciais de negócio:**

- **Privacidade:** processamento 100% local, nada sai da máquina/rede do usuário.
- **Tempo real:** o texto aparece enquanto a pessoa fala.
- **Multi-idioma:** detecção automática ou escolha entre 9 idiomas.
- **Identificação de quem fala:** distingue o próprio usuário, o áudio do sistema e, quando
  possível, participantes individuais.
- **Pós-processamento por IA:** correção de erros e geração de resumos com tópicos e próximos
  passos, sem trabalho manual.

---

## 2. Personas / Usuários

Use estas personas no campo **"Como [persona]"** das histórias:

- **Profissional de reuniões** — registra reuniões e conversas para não perder informações e
  ter um resumo pronto depois. Persona principal.
- **Equipe com requisito de privacidade** — precisa transcrever conteúdo sensível sem usar
  serviços de nuvem.
- **Administrador/operador** — mantém o serviço no ar e acompanha o uso do banco de dados.

> Observação: cada pessoa tem sua própria conta (usuário/senha) e vê apenas suas próprias
> transcrições, correções, resumos, instruções e tarefas — dados são isolados por conta
> (multi-tenant). Ver seção 4.5.

---

## 3. Glossário de Domínio

Termos a serem usados de forma **consistente** em histórias e critérios de aceite:

| Termo | Significado de negócio |
|---|---|
| **Transcrição (bruta)** | Texto gerado diretamente do áudio, sem revisão. |
| **Transcrição corrigida** | Versão da transcrição revisada pela IA (gramática, pontuação, nomes). |
| **Sumarização / Resumo** | Documento estruturado com título, descrição, tópicos e próximos passos, gerado a partir de uma transcrição. |
| **Chunk** | Tamanho do trecho de áudio (em segundos) acumulado antes de transcrever. Trade-off entre rapidez de resposta e precisão. |
| **Motor de transcrição (STT)** | Componente que converte áudio em texto em tempo real. O Transcriber suporta três motores: _Whisper_, _Voxtral_ e _Qwen2-Audio_, cada um com seus próprios modelos (ver "Modelos"). |
| **Modelos (STT)** | Os tamanhos disponíveis por motor: _Whisper_ (tiny → large-v3), _Voxtral_ (Mini 3B, Small 24B) e _Qwen2-Audio_ (7B, 7B-Instruct). Motor e modelo são escolhidos uma vez na gravação. |
| **Diarização** | Recurso opcional que identifica diferentes participantes da conversa. |
| **Rótulos de locutor** | Marcações de quem falou: `EU` (microfone do usuário), `OUTRO` (áudio do sistema), `PESSOA 1…N` (participantes identificados pela diarização). |
| **Job / processamento assíncrono** | Tarefa de correção ou sumarização que roda em segundo plano, com acompanhamento de progresso ao vivo. |
| **Fonte de áudio** | Origem do som: modo dispositivo (microfone + áudio do sistema) ou modo navegador (captura pelo navegador). |
| **Conta** | Identidade de um usuário (usuário + senha), criada por autocadastro. Todas as transcrições, correções, resumos, instruções e tarefas pertencem a uma conta e são isoladas das demais. |

---

## 4. Mapa de Funcionalidades

A interface tem **4 áreas** principais, navegáveis pelo topo: **Gravar**, **Transcrições**,
**Corrigidas** e **Sumarizadas**, além de **Login** e **Criar conta**.

### 4.1 Gravar

Tela principal para iniciar uma sessão de transcrição ao vivo.

- **Título da reunião** (obrigatório) — identifica a transcrição que será criada.
- **Fonte de áudio:**
  - *Modo dispositivo* (padrão): o usuário escolhe o **microfone** (entrada) e o **áudio do
    sistema** (saída) em listas separadas.
  - *Modo navegador* (alternativo): captura o áudio pelo próprio navegador, com botões para
    ativar **microfone** e **áudio do sistema**, cada um com indicador "Ativo / Não ativo".
- **Modelo Whisper** — escolha entre opções de mais rápido/menos preciso até mais lento/mais
  preciso.
- **Tamanho do chunk (segundos)** — opções de mais responsivo (menor) a mais preciso (maior).
- **Idioma do áudio** — auto-detectar ou um idioma específico (Português, Inglês, Espanhol,
  Francês, Alemão, Italiano, Japonês, Chinês).
- **Iniciar / Parar** — inicia a gravação (só habilita com título preenchido) e a encerra.
- **Transcrição ao vivo** — o texto aparece em tempo real, com horário e rótulo de locutor.
- **Cronômetro** — mostra o tempo decorrido da sessão.
- **Indicadores de status** — sinalizam o estado atual (aguardando, carregando modelo,
  transcrevendo, parando).
- **Limpar** — limpa a área de transcrição ao vivo.

Ao encerrar, a sessão é salva como uma nova **transcrição bruta**.

### 4.2 Transcrições

Lista das transcrições brutas salvas.

- **Listagem** com: título, data, tamanho.
- **Atualizar** a lista.
- Ações por item:
  - **Visualizar** (somente leitura)
  - **Editar** (ajuste manual do texto e salvar)
  - **Deletar** (com confirmação)
  - **Corrigir** → inicia um job de correção por IA
  - **Sumarizar** → inicia um job de resumo por IA
- **Estado vazio:** mensagem indicando que não há transcrições.

### 4.3 Corrigidas

Lista das transcrições já corrigidas pela IA.

- Mesma listagem (título, data, tamanho) e ações de **Visualizar**, **Editar**, **Deletar**.
- Ação de **Sumarizar** (não há "Corrigir" novamente).
- **Estado vazio:** mensagem indicando que não há transcrições corrigidas.

### 4.4 Sumarizadas

Lista dos resumos gerados.

- Listagem e ações de **Visualizar**, **Editar**, **Deletar**.
- **Formato do resumo:**
  - **Título** da conversa
  - **Descrição** (resumo em poucas frases)
  - **Tópicos discutidos** — cada um com horário, nome do tópico e pontos-chave
  - **Próximos passos / ações** — itens de ação, com responsável quando identificado
- **Estado vazio:** mensagem indicando que não há sumarizações.

### 4.5 Autenticação e contas

- **Criar conta** — autocadastro público (usuário + senha, sem aprovação manual). Qualquer
  pessoa com acesso ao app pode criar sua própria conta.
- **Login** com usuário e senha.
- **Logout** ("Sair") disponível no topo quando autenticado.
- As áreas internas só são acessíveis após autenticação (quando o login está habilitado).
- **Isolamento por conta:** cada usuário só vê e gerencia suas próprias transcrições,
  correções, resumos, instruções e tarefas. Duas pessoas podem gravar reuniões
  simultaneamente, cada uma na sua própria sessão, sem interferir uma na outra.

---

## 5. Fluxos de Negócio (jornadas do usuário)

**Jornada principal:**

1. **Gravar** — preenche título, escolhe fonte de áudio/modelo/idioma e inicia.
2. **(Opcional) Visualizar/Editar** — revisa ou ajusta a transcrição bruta.
3. **Corrigir** — gera a versão corrigida pela IA.
4. **Sumarizar** — gera o resumo estruturado (a partir da bruta ou da corrigida).

**Processamento assíncrono (jobs de Corrigir/Sumarizar):**

- O usuário escolhe o modelo de IA e inicia o processamento.
- Acompanha o **progresso ao vivo** (log da execução) em um painel.
- Ao final, vê o estado de **conclusão** (com o arquivo gerado) ou de **erro**.
- Pode **cancelar** um processamento em andamento.
- Se a página for recarregada, o acompanhamento de um job em andamento é **retomado**.

---

## 6. Regras de Negócio

- **Título obrigatório:** não é possível iniciar a gravação sem informar o título da reunião.
- **Bloqueio de reprocessamento:** se já existe um arquivo de saída para aquela correção/resumo,
  o reprocessamento é bloqueado (evita sobrescrever resultado existente).
- **Fallback de fonte de áudio:** quando não há dispositivos de áudio disponíveis, o produto
  oferece o modo de captura pelo navegador.
- **Diarização é opcional:** sem o recurso de identificação de participantes, todo o áudio do
  sistema é rotulado como `OUTRO`; com ele, os participantes aparecem como `PESSOA 1…N`.
- **Confirmação ao deletar:** excluir uma transcrição, correção ou resumo exige confirmação.
- **Origem do resumo:** a sumarização pode partir da transcrição bruta ou da corrigida.

---

## 7. Estados e Mensagens

Úteis para escrever **critérios de aceite verificáveis**.

| Situação | O que o usuário vê |
|---|---|
| Sem transcrições / corrigidas / resumos | Mensagem de lista vazia correspondente |
| Lista carregando | Indicador de carregamento |
| Gravação aguardando título | Aviso para preencher o título antes de iniciar |
| Carregando modelo | Status "Carregando modelo…" |
| Transcrevendo | Status ativo + cronômetro + indicador pulsante |
| Job em execução | Selo "Em execução…" + log ao vivo |
| Job concluído | Selo de concluído + nome do arquivo salvo |
| Job com erro | Selo de erro + mensagem |
| Reprocessamento bloqueado | Aviso de que o arquivo já existe |
| Sem modelos de IA disponíveis | Aviso para verificar a configuração do serviço de IA |
| Credenciais inválidas no login | Mensagem de erro de autenticação |
| Falha de comunicação | Aviso de erro de comunicação com o servidor |

---

## 8. Fora de Escopo / Não-funcionalidades

O produto **hoje não** possui (e histórias não devem assumir que existem, salvo se a própria
história pedir para **criar** o recurso):

- **Recuperação/reset de senha** — quem esquece a senha não tem autoatendimento hoje.
- **Perfis, papéis ou permissões diferenciadas entre usuários** (ex: admin vs. membro) — toda
  conta autenticada tem as mesmas capacidades sobre seus próprios dados.
- **Cobrança, planos, assinaturas, cotas ou créditos** — todas as funções são livres para quem
  tem conta.
- **Notificações, e-mails ou webhooks** — todo feedback é dado na própria interface.
- **Compartilhamento externo / colaboração em tempo real entre usuários** — cada conta só
  acessa seus próprios dados; não há compartilhamento de transcrições entre contas.

> Deixar isso explícito evita que histórias sejam geradas para funcionalidades que não fazem
> parte do produto atual.

---

## 9. Conexão com o fluxo de histórias

Este documento é a referência de negócio para as skills **`/story`** e **`/work`**:

- **`/story`** deve usar as **personas** (seção 2), o **glossário** (seção 3) e as **regras**
  (seções 6 e 8) para escrever histórias e critérios de aceite consistentes.
- Cada história nasce em **BACKLOG** e percorre o fluxo do board:
  `BACKLOG → PLANNING → IN PROGRESS → TESTING → RELEASE → COMPLETE`.
- Quando uma história tocar uma funcionalidade do **Mapa de Funcionalidades** (seção 4),
  referenciar a área da UI correspondente (Gravar, Transcrições, Corrigidas, Sumarizadas,
  Autenticação) para situar a entrega.

---

## 10. Base de Conhecimento (RAG / LightRAG)

Além desta documentação de **negócio**, o projeto possui uma **base de conhecimento técnica**
indexada com LightRAG — um grafo de conhecimento construído a partir do **código, docs e
configs** do repositório. Use-a para entender **como o código realmente funciona** (arquivos,
módulos, dependências, relações entre componentes) — algo que o PRODUTO.md não cobre.

**Quando consultar:**
- Para localizar onde uma funcionalidade está implementada no código.
- Para entender dependências e impacto de uma mudança antes de planejar.
- Para responder dúvidas técnicas sobre a arquitetura sem ter que varrer o repositório inteiro.

**Como consultar:**

A forma preferida é via **tools MCP** (servidor `lightrag`, registrado em `.mcp.json`):

| Tool MCP | Para quê |
|---|---|
| `kg_query` | Pergunta em linguagem natural sobre o código/projeto (busca híbrida no grafo). |
| `kg_stats` | Panorama da base: nº de entidades, relações, documentos indexados. |
| `kg_insert_text` | Adicionar uma nota/decisão à base de conhecimento. |

Se as tools MCP não estiverem disponíveis na sessão, use a CLI equivalente:

```bash
uv run --project tools/lightrag rag search "como funciona X"   # busca híbrida
uv run --project tools/lightrag rag ask    "explique Y"
uv run --project tools/lightrag rag stats                       # panorama da base
```

**Onde fica:** o storage do grafo está em `tools/lightrag/rag_storage/` e o export
navegável (Obsidian) em `docs/knowledge-graph/`.

> ⚠️ **Não** rode `rag index` / `kg-index` dentro da sessão do agente (indexação é pesada e
> deve ser feita no terminal do usuário). As tools MCP expostas (`kg_query`, `kg_stats`,
> `kg_insert_text`) são seguras para uso pelo agente.
