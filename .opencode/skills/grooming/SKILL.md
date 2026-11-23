---
name: grooming
description: Use when asked to do grooming, technically refine a task, run the grooming process, or process a ready-to-groom GitHub issue. Triggers on phrases like "fazer grooming", "refinar a task", "grooming da issue", "processar grooming", "technical grooming", or any request to transform a business-analyzed issue into a technically ready one. Always invoke this skill for grooming work — never handle it manually.
---

# grooming

## Overview

Conduz o grooming técnico de uma issue do GitHub: lê a análise de negócio já existente na descrição e comentários, gera as **Definições Técnicas** necessárias para o desenvolvimento, insere o resultado na descrição da issue e a move para o status de "pronto para desenvolver".

Este fluxo complementa o `business-analysis` — onde aquele foca em *o quê* e *por quê*, este foca em *como* implementar.

## Configuração do Projeto

Esta skill é agnóstica a projeto. Antes de qualquer chamada ao GitHub, leia `.opencode/dev-skills.config.json` na raiz do projeto.

> **Se o arquivo não existir, ou não tiver a seção `github`:** informe o usuário — *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* — e encerre sem executar nenhuma ação. Nunca invente repositório, labels de status ou de prioridade.

| Referência nesta skill | Chave no config |
|---|---|
| Repositório das issues | `github.repo` |
| Labels de status | `statuses.*` |
| Labels de prioridade | `priorities.*` |

## Operações no GitHub

Todas as operações usam a `gh` CLI, sempre com `-R {{github.repo}}`. Se o `gh` não estiver autenticado, oriente o usuário a rodar `gh auth login` e encerre.

| Operação | Comando |
|---|---|
| Ler issue | `gh issue view <n> -R {{github.repo}} --json number,title,body,labels,state,createdAt,url` |
| Ler comentários | `gh issue view <n> -R {{github.repo}} --json comments` |
| Listar por status | `gh issue list -R {{github.repo}} --state open --label "<label de status>" --json number,title,labels,createdAt --limit 100` |
| Mudar de status | `gh issue edit <n> -R {{github.repo}} --remove-label "<label de status atual>" --add-label "<novo label>"` |
| Anexar à descrição | escrever `body atual + "\n\n---\n\n" + artefato` em um arquivo temporário → `gh issue edit <n> -R {{github.repo}} --body-file <arquivo>` |

> **Label de status atual** = o label da issue que corresponde a algum valor de `statuses.*`. Se a issue tiver mais de um label de status (estado inconsistente), remova todos na troca e avise o usuário.

## Fluxo Obrigatório

```
[início]  → Verificar .opencode/dev-skills.config.json (se ausente: orientar rodar project-setup e encerrar)
         → Identificar issue → Mover para in_grooming
         → Ler descrição + comentários
         → Gerar Definições Técnicas
         → Atualizar descrição da issue (append)
         → Mover para ready_to_develop
[fim]
```

**Nunca pule etapas. A ordem é obrigatória.**

---

## Passo 1 — Identificar a Issue

**Se um número ou URL de issue foi passado como argumento:** use-o diretamente via `gh issue view`.

**Se nenhuma issue foi especificada:**
1. Listar as issues abertas com o label `{{statuses.ready_to_grooming}}` (ver "Listar por status").
2. Ordenar o resultado pela tabela de criticidade abaixo (mais crítico primeiro), a partir do label de prioridade presente em `labels`. Dentro da mesma prioridade, ordenar por `createdAt` ASC (issue mais antiga primeiro).
3. Selecionar o primeiro item da lista ordenada.
4. Confirmar com o usuário: *"Encontrei a issue #[número] '[título]' (prioridade: [prioridade]). Vou iniciar o grooming técnico. Confirma?"*

**Tabela de criticidade — labels de prioridade (ordem decrescente):**

| Ordem | Prioridade (semântica) | Chave no config |
|------|--------------------|------|
| 1 | Urgente | `priorities.urgent` |
| 2 | Alta | `priorities.high` |
| 3 | Normal | `priorities.normal` |
| 4 | Baixa | `priorities.low` |
| 5 | (sem label de prioridade) | — |

> Se o projeto não configurou `priorities`, ordene apenas por `createdAt` ASC.

**Se não houver issues em Ready To Grooming:** informar o usuário e encerrar.

---

## Passo 2 — Mover para In Grooming

Trocar o label de status para `{{statuses.in_grooming}}` (ver "Mudar de status") antes de qualquer análise.

---

## Passo 3 — Ler Contexto da Issue

Executar em paralelo:
- `gh issue view <n> --json number,title,body,labels` — para obter título, descrição completa (incluindo a Análise de Negócio gerada pela skill `business-analysis`, se houver) e labels da issue
- `gh issue view <n> --json comments` — para obter comentários adicionais (discussões e esclarecimentos que não fazem parte do artefato principal)

Com o conteúdo coletado, extraia mentalmente:
- O **problema de negócio** e o **objetivo** esperado
- As **regras de negócio** e **critérios de aceitação**
- O **escopo** definido (o que está dentro e fora)
- Quaisquer **restrições** ou **riscos** levantados

---

## Passo 4 — Gerar as Definições Técnicas

Com base no contexto extraído, produza o documento de Definições Técnicas em **português brasileiro**. Seja preciso e acionável — este documento será a referência do desenvolvedor durante a implementação.

Use o template abaixo. Adapte as seções ao que for relevante para a issue; se uma seção não se aplica, escreva `N/A` com uma breve justificativa.

```markdown
## Definições Técnicas

**Issue:** #[número] [título]
**Grooming por:** [usuário logado — `gh api user --jq .login`]
**Data:** [data atual]

### Visão Geral Técnica
[Resumo de 2-4 linhas sobre a abordagem técnica adotada, conectando diretamente ao objetivo de negócio.]

### Entidades Envolvidas
Liste as entidades de domínio (tabelas, modelos, agregados) afetadas ou criadas:
- **[Entidade]:** [descrição do papel e campos relevantes a adicionar/modificar]

### Endpoints / Serviços
Liste os endpoints REST ou serviços a criar/modificar. Para cada um:
- **[MÉTODO] /caminho** — [descrição do que faz]
  - Request: [campos principais]
  - Response: [campos principais]
  - Autenticação: [sim/não, tipo]

### Mudanças no Banco de Dados
- Novas tabelas: [lista]
- Colunas adicionadas: [tabela.coluna — tipo — descrição]
- Índices necessários: [tabela(coluna) — motivo]
- Migrations: [descrever brevemente o que cada migration fará]

### Regras e Validações Técnicas
Derivadas das regras de negócio — traduza cada regra para a validação que o código deve aplicar:
1. [Regra técnica com referência à regra de negócio de origem]
2. ...

### Casos de Borda
Situações que o desenvolvedor deve tratar explicitamente:
- [Caso]: [comportamento esperado]

### Critérios de Aceite Técnicos
Verificações que o desenvolvedor pode fazer antes de abrir o PR:
- [ ] [Critério objetivo e verificável]
- [ ] [Critério objetivo e verificável]

### Dependências e Ordem de Implementação
Se houver dependências entre partes da solução, indique a ordem sugerida:
1. [Primeiro passo]
2. [Segundo passo]

### Perguntas Técnicas em Aberto
[Se houver ambiguidades técnicas não resolvíveis apenas com o contexto disponível, liste aqui para o desenvolvedor resolver antes de começar.]
```

---

## Passo 5 — Atualizar a Descrição da Issue

Construa a nova descrição **preservando integralmente o conteúdo original** e **adicionando** as Definições Técnicas ao final, separadas por uma linha horizontal (`---`). Grave-a conforme "Anexar à descrição":
```
[descrição original da issue]

---

[documento de Definições Técnicas gerado no Passo 4]
```

> Nunca sobrescreva a descrição original — apenas acrescente. O histórico de negócio deve ser preservado para referência futura. Se o corpo resultante passar do limite do GitHub (65.536 caracteres), publique o artefato como comentário (`gh issue comment <n> -R {{github.repo}} --body-file <arquivo>`) e anexe à descrição apenas o link para esse comentário.

---

## Passo 6 — Mover para Ready To Develop

Trocar o label de status de `{{statuses.in_grooming}}` para `{{statuses.ready_to_develop}}`.

Confirmar ao usuário: *"Grooming técnico concluído. A issue #[número] '[título]' foi movida para Ready To Develop com as Definições Técnicas adicionadas à descrição."*

---

## Princípios do Grooming

| Princípio | Aplicação |
|-----------|-----------|
| Baseado no negócio | Cada decisão técnica deve ser rastreável a uma regra ou objetivo de negócio |
| Acionável | O desenvolvedor deve poder implementar com o documento, sem perguntas adicionais |
| Sem ambiguidade | Prefira ser explícito a deixar espaço para interpretação |
| Preservar histórico | Nunca apagar a descrição original — sempre adicionar ao final |
| Foco no desenvolvedor | Escreva como se fosse uma spec que você mesmo usaria para implementar |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Sobrescrever a descrição original | Sempre preserve o conteúdo existente e adicione ao final |
| Gerar definições genéricas | Ancore cada item em detalhe específico da análise de negócio lida |
| Pular Passo 2 (mover para In Grooming) | O label de status deve refletir o estado real da issue em tempo real |
| Adicionar o novo label de status sem remover o anterior | Sempre trocar o label em uma única chamada (`--remove-label` + `--add-label`) |
| Deixar seções vazias sem justificativa | Se não se aplica, explique brevemente o porquê |
| Mover para Ready To Develop antes de atualizar a descrição | A descrição deve ser atualizada antes da mudança de status final |
| Executar sem `.opencode/dev-skills.config.json` | Rodar `project-setup` primeiro — nunca inventar repositório ou labels |
