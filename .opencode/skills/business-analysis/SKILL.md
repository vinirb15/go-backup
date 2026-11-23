---
name: business-analysis
description: Use when asked to analyze, scope, or refine a GitHub issue from a business perspective — covers requirements elicitation, acceptance criteria, stakeholder mapping, business rules, and readiness for grooming. Do NOT use for technical design or implementation decisions.
---

# business-analysis

## Overview

Conduz a análise de negócio de uma issue do GitHub de forma autônoma: infere contexto, objetivos, stakeholders e critérios de aceitação a partir da descrição da issue, comentários e issues relacionadas. Só interage com o usuário quando uma dúvida é crítica e não pode ser resolvida com o contexto disponível. Não aborda aspectos técnicos de implementação.

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
         → Buscar issue → Mover para statuses.scoping
         → Conduzir análise de negócio (autônomo — pergunta só se dúvida crítica)
         → Escrever artefato final → Adicionar à descrição da issue
         → Mover para statuses.ready_to_grooming
[fim]
```

**Nunca pule etapas. A ordem é obrigatória.**

---

## Passo 1 — Identificar a Issue

**Se um número ou URL de issue foi passado como argumento:** use-o diretamente via `gh issue view`.

**Se nenhuma issue foi especificada:**
1. Listar as issues abertas com o label `{{statuses.backlog}}` (ver "Listar por status").
2. Ordenar o resultado pela tabela de criticidade abaixo (mais crítico primeiro), a partir do label de prioridade presente em `labels`. Dentro da mesma prioridade, ordenar por `createdAt` ASC (issue mais antiga primeiro).
3. Selecionar o primeiro item da lista ordenada.
4. Informar o usuário: *"Selecionei a issue #[número] '[título]' (prioridade: [prioridade]) para análise."* — e seguir direto para o Passo 2, sem esperar confirmação.

**Exceção:** se dois ou mais candidatos empatarem tanto em prioridade quanto em `createdAt`, a escolha entre eles é uma dúvida crítica — perguntar ao usuário qual analisar antes de prosseguir.

**Tabela de criticidade — labels de prioridade (ordem decrescente):**

| Ordem | Prioridade (semântica) | Chave no config |
|------|--------------------|------|
| 1 | Urgente | `priorities.urgent` |
| 2 | Alta | `priorities.high` |
| 3 | Normal | `priorities.normal` |
| 4 | Baixa | `priorities.low` |
| 5 | (sem label de prioridade) | — |

> Se o projeto não configurou `priorities`, ordene apenas por `createdAt` ASC.

**Se não houver issues em backlog:** informar o usuário e encerrar.

---

## Passo 2 — Mover para Scoping

Trocar o label de status para `{{statuses.scoping}}` (ver "Mudar de status") antes de qualquer análise.

---

## Passo 3 — Conduzir a Análise de Negócio

Conduza a análise de forma **autônoma**, cobrindo **todas** as seções abaixo com base no contexto levantado. Interaja com o usuário apenas quando restar uma dúvida crítica (ver 3.0 e 3.8) — nunca seção por seção.

### 3.0 Levantamento de Contexto

Antes de preencher qualquer seção, levante o máximo de contexto disponível:

- Descrição completa da issue (obtida no Passo 1, ou via `gh issue view` se ainda incompleta)
- Comentários da issue (`gh issue view <n> --json comments`)
- Issues vinculadas ou mencionadas no corpo/comentários (`#<número>`), se houver — leia-as com `gh issue view`
- Conhecimento de domínio já estabelecido (padrões de negócio recorrentes em outras issues do mesmo repositório, se identificáveis)

Use essas fontes para preencher o máximo possível das seções 3.1–3.7 sem perguntar. Só o que não puder ser inferido a partir daqui — e for crítico — vira pergunta no Passo 3.8.

**Critério de criticidade — quando assumir vs. quando perguntar:**

| Cenário | Ação |
|---|---|
| Informação ausente, mas inferível do contexto levantado (descrição, comentários, issues relacionadas, domínio já conhecido) | Assumir com base no contexto e registrar em **Premissas** |
| Detalhe de menor impacto (redação, nome de stakeholder secundário, prioridade entre riscos menores) | Assumir a interpretação mais razoável e seguir |
| Ambiguidade que muda o escopo da entrega, ou tem impacto legal/financeiro/compliance/cliente | Perguntar ao usuário — dúvida crítica |
| Duas ou mais interpretações plausíveis e mutuamente exclusivas, sem sinal claro no contexto de qual está correta | Perguntar ao usuário — dúvida crítica |

### 3.1 Contexto e Problema

- Qual é o problema de negócio que esta issue resolve?
- O que acontece hoje que gera dor ou ineficiência?
- Há alguma regulação, contrato ou prazo que torne isso urgente?

### 3.2 Objetivo de Negócio

- Qual é o resultado esperado quando esta issue estiver resolvida?
- Como saberemos que o objetivo foi alcançado? (métricas, indicadores)
- Há um objetivo secundário ou benefício adjacente esperado?

### 3.3 Stakeholders

- Quem são os usuários diretos desta funcionalidade/mudança?
- Quem será impactado indiretamente?
- Quem é o responsável por aprovar a entrega (dono do produto, cliente, área de negócio)?

### 3.4 Regras de Negócio

- Há restrições, políticas internas ou regras externas que se aplicam?
- Há casos especiais ou exceções que devem ser tratados?
- Existe alguma dependência com outro processo ou sistema de negócio?

### 3.5 Critérios de Aceitação

Formule ao menos 3 critérios no formato **"Dado [contexto], quando [ação], então [resultado esperado]"**, com base no contexto levantado em 3.0. Só volte a um critério com o usuário se ele se enquadrar como dúvida crítica na tabela de 3.0.

### 3.6 Escopo

- O que está **dentro** do escopo desta issue?
- O que está **fora** do escopo (explicitamente)?
- Há itens futuros relacionados que não devem ser tratados agora?

### 3.7 Riscos e Premissas

- Quais premissas estamos assumindo que, se falsas, invalidariam a solução?
- Quais riscos de negócio existem (impacto no cliente, compliance, receita)?
- Há dependências externas (terceiros, outras equipes) que podem bloquear?

### 3.8 Consolidar Dúvidas Críticas (se houver)

Depois de preencher 3.1–3.7, revise se sobrou algum ponto classificado como "dúvida crítica" na tabela de 3.0. Se sim, envie **uma única mensagem** ao usuário consolidando todas as dúvidas de uma vez — nunca uma pergunta por seção. Se não sobrou nenhuma, siga direto para o Passo 4 sem interação alguma.

---

## Passo 4 — Escrever o Artefato Final

Ao término da análise, produza um documento estruturado com as seções abaixo. Escreva em **português**, com linguagem de negócio — sem jargões técnicos de implementação.

Todo conteúdo deve ser estruturado em **lista alternada**: campos e itens alternam entre marcadores `-` e `1.`, sem blocos de prosa.

```markdown
## Análise de Negócio

- **Issue:** #[número] [título]
1. **Analista:** [usuário logado — `gh api user --jq .login`]
- **Data:** [data atual]

### Problema
- **Situação atual:** [o que acontece hoje que gera dor]
1. **Impacto:** [consequência para o negócio ou usuário]
- **Urgência:** [prazo, regulação ou evento que torna urgente — se houver]

### Objetivo
1. **Resultado esperado:** [o que se quer alcançar]
- **Métrica de sucesso:** [como saberemos que foi alcançado]
1. **Benefício secundário:** [benefício adjacente — se houver]

### Stakeholders
- **Usuários diretos:** [lista]
1. **Impactados indiretamente:** [lista]
- **Responsável pela aprovação:** [nome/área]

### Regras de Negócio
- [Regra 1]
1. [Regra 2]
- [Regra 3]
...

### Critérios de Aceitação
- [ ] Dado [contexto], quando [ação], então [resultado]
- [ ] Dado [contexto], quando [ação], então [resultado]
- [ ] Dado [contexto], quando [ação], então [resultado]

### Escopo
- **Dentro:** [itens incluídos]
1. **Fora:** [itens explicitamente excluídos]
- **Futuro:** [itens relacionados para entregas posteriores]

### Riscos e Premissas
- **Premissa:** [premissa assumida autonomamente a partir do contexto — não perguntada ao usuário]
1. **Risco:** [risco e impacto estimado]
- **Dependência externa:** [terceiro ou outra equipe que pode bloquear — se houver]

### Perguntas em Aberto
- [Dúvida crítica levantada no Passo 3.8 que não foi respondida pelo usuário — omitir seção se não houver]
```

---

## Passo 5 — Atualizar a Descrição da Issue

Obtenha o `body` atual da issue via `gh issue view <n> --json body` (caso o Passo 1 não o tenha retornado — por exemplo, quando a issue foi selecionada via `gh issue list`).

Grave a nova descrição (ver "Anexar à descrição") contendo:
```
[descrição atual da issue]

---

[documento de Análise de Negócio gerado no Passo 4]
```

> Nunca sobrescreva a descrição existente — sempre acrescente ao final, separado por `---`. Se o corpo resultante passar do limite do GitHub (65.536 caracteres), publique o artefato como comentário (`gh issue comment <n> -R {{github.repo}} --body-file <arquivo>`) e anexe à descrição apenas o link para esse comentário.

---

## Passo 6 — Mover para Ready To Grooming

Trocar o label de status de `{{statuses.scoping}}` para `{{statuses.ready_to_grooming}}`.

---

## Princípios da Análise

| Princípio | Aplicação |
|-----------|-----------|
| Inferir primeiro | Usar o contexto disponível (descrição, comentários, issues relacionadas) por padrão; perguntar somente diante de dúvida crítica (ver tabela de criticidade no Passo 3.0) |
| Foco em negócio | Não mencione tecnologia, banco de dados, APIs ou arquitetura |
| Critério mensurável | Todo objetivo deve ter forma de validação observável |
| Escopo explícito | O que está fora é tão importante quanto o que está dentro |
| Linguagem do negócio | Use termos do domínio do cliente, não do desenvolvedor |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Perguntar sem antes levantar o contexto disponível | Levantar contexto (3.0) primeiro — perguntar é exceção, não padrão |
| Perguntar seção por seção | Consolidar dúvidas críticas em uma única rodada ao final do Passo 3 (3.8) |
| Critérios vagos ("funcionar bem") | Reformular com Dado/Quando/Então |
| Mencionar detalhes técnicos | Redirecionar ao impacto de negócio |
| Adicionar o novo label de status sem remover o anterior | Sempre trocar o label em uma única chamada (`--remove-label` + `--add-label`) |
| Mover para Ready To Grooming antes de atualizar a descrição | A descrição deve ser atualizada antes da mudança de status final |
| Executar sem `.opencode/dev-skills.config.json` | Rodar `project-setup` primeiro — nunca inventar repositório ou labels |
