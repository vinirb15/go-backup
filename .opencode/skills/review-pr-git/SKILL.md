---
name: review-pr-git
description: Use when asked to review the diff of a branch, do code review straight from git (no PR), or process a GitHub issue that is "in review" from the local git diff instead of a pull request. Triggers on phrases like "revisar a branch", "code review com git", "revisar o diff da branch", "review da task sem gh", "revisão da branch", or any request to review the code of a developed task from the local git diff and decide if it advances to testing. Use this variant when there is no PR to review (branch only) or when you want to review straight from git. Always invoke a review skill for review work — never handle it manually.
---

# review-pr-git

## Overview

Conduz a revisão de código de uma issue do GitHub que está em revisão, obtendo o código a revisar **apenas via `git`** (sem depender de um Pull Request). Lê as **Definições Técnicas** geradas no grooming, descobre a branch da issue, revisa o diff da branch contra a branch base, anexa um relatório de revisão à descrição da issue e move o status conforme o veredito — **Aprovado → Testing** ou **Reprovado → In Develop**.

É a variante "git puro" da `review-pr-gh`. Ambas seguem o mesmo fluxo e produzem o mesmo relatório; a diferença é **a fonte do código a revisar**:

| | Fonte do diff | Pré-requisito | Contexto extra |
|---|---|---|---|
| `review-pr-gh` | PR no GitHub (`gh pr diff`) | `gh` instalado e autenticado | Metadados do PR (autor, descrição, estado) |
| **`review-pr-git`** | Diff da branch (`git diff`) | `git` + remote `origin` (o `gh` continua necessário para ler/atualizar a issue) | `git log` da branch (commits) |

Use esta skill quando a branch ainda não tem PR aberto, ou quando preferir revisar direto do git. A referência central da revisão são as **Definições Técnicas na issue** — não a descrição de um PR — então a ausência dos metadados do PR não compromete a revisão.

## Configuração do Projeto

Esta skill é agnóstica a projeto. Antes de qualquer chamada ao GitHub ou ao repositório, leia `.opencode/dev-skills.config.json` na raiz do projeto.

> **Se o arquivo não existir, ou não tiver a seção `github`:** informe o usuário — *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* — e encerre sem executar nenhuma ação. Nunca invente repositório, labels de status ou caminhos de pacote.

| Referência nesta skill | Chave no config |
|---|---|
| Repositório das issues | `github.repo` |
| Labels de status | `statuses.*` |
| Branch base | `repo.default_branch` |
| Backend / frontend / shared | `repo.packages.{backend,frontend,shared}` (ausentes → pule a referência) |
| ORM / ferramenta de migration | `repo.orm` (se `"none"`, pule a checagem de migrations) |
| Config de cobertura | `repo.coverage_config_path` (se ausente, pule a checagem de threshold) |

## Operações no GitHub

Todas as operações de issue usam a `gh` CLI, sempre com `-R {{github.repo}}`. Se o `gh` não estiver autenticado, oriente o usuário a rodar `gh auth login` e encerre.

| Operação | Comando |
|---|---|
| Ler issue | `gh issue view <n> -R {{github.repo}} --json number,title,body,labels,state,createdAt,url` |
| Ler comentários | `gh issue view <n> -R {{github.repo}} --json comments` |
| Listar por status | `gh issue list -R {{github.repo}} --state open --label "<label de status>" --json number,title,labels,createdAt --limit 100` |
| Mudar de status | `gh issue edit <n> -R {{github.repo}} --remove-label "<label de status atual>" --add-label "<novo label>"` |
| Anexar à descrição | escrever `body atual + "\n\n---\n\n" + relatório` em um arquivo temporário → `gh issue edit <n> -R {{github.repo}} --body-file <arquivo>` |

> **Label de status atual** = o label da issue que corresponde a algum valor de `statuses.*`. Se a issue tiver mais de um label de status (estado inconsistente), remova todos na troca e avise o usuário.

## Fluxo Obrigatório

```
[início]  → Verificar .opencode/dev-skills.config.json (se ausente: orientar rodar project-setup e encerrar)
         → Identificar issue (in_review)
         → Ler descrição + comentários (Definições Técnicas do grooming)
         → Descobrir a branch da issue
         → git fetch + git diff <branch-base>...branch + git log
         → Revisar o diff contra os critérios de aceite técnicos
         → Escrever o Relatório de Revisão → Adicionar à descrição da issue
         → Veredito: Aprovado → testing | Reprovado → in_develop
[fim]
```

**Nunca pule etapas. A ordem é obrigatória.**

> Diferente das skills de análise/grooming, a issue **já entra** nesta skill no status "In Review" — portanto **não há mudança de status no início**. A primeira ação é ler o contexto da issue.

---

## Passo 1 — Identificar a Issue

**Se um número/URL de issue ou um nome de branch foi passado como argumento:** use-o diretamente.
- Número/URL de issue → `gh issue view`.
- Nome de branch → use a branch direto e localize a issue vinculada pelo número no nome da branch (`feat/<número>-...`) (ver Passo 3).

**Se nada foi especificado:**
1. Listar as issues abertas com o label `{{statuses.in_review}}` (ver "Listar por status").
2. Se houver múltiplas issues, listar número + título e perguntar ao usuário qual deseja revisar.
3. Se houver apenas uma, confirmar: *"Encontrei a issue #[número] '[título]' em In Review. Vou iniciar a revisão da branch. Confirma?"*

**Se não houver issues em In Review:** informar o usuário e encerrar.

---

## Passo 2 — Ler Contexto da Issue

Executar em paralelo:
- `gh issue view <n> --json number,title,body,labels` — para obter título, descrição completa (que já reúne a **Análise de Negócio** e as **Definições Técnicas** do grooming) e labels da issue.
- `gh issue view <n> --json comments` — para obter comentários adicionais (discussões e qualquer referência à branch que não faça parte dos artefatos principais).

Com o conteúdo coletado, extraia mentalmente:
- Os **Critérios de Aceite Técnicos** definidos no grooming (a referência principal da revisão).
- As **Entidades**, **Endpoints/Serviços** e **Mudanças no Banco** previstos.
- As **Regras e Validações Técnicas** e os **Casos de Borda** que o código deveria tratar.
- O **objetivo de negócio** original (para checar se a implementação não desviou do propósito).

---

## Passo 3 — Descobrir a Branch da Issue

Como não há PR envolvido, a unidade revisada é a **branch**. Descubra-a nesta ordem:
1. nome da branch citado nos **comentários** ou na **descrição** da issue;
2. liste as branches remotas com `git branch -r` e correlacione pelo número da issue (a skill `develop` cria `feat/<número>-<slug>` ou `fix/<número>-<slug>`);
3. se nada for conclusivo, **pergunte ao usuário** qual é a branch. Nunca adivinhe.

> Esta skill depende de um remote `origin` configurado. Se `git` não tiver remote ou a branch não existir no `origin`, informe o usuário e encerre sem mover o status.

---

## Passo 4 — Obter o Diff da Branch

Atualize as refs remotas e colete o diff e o histórico (a base é `repo.default_branch`; ajuste se a branch não derivar dela):

```bash
git fetch origin
git diff origin/{{repo.default_branch}}...origin/<branch>        # diff: só o que a branch adicionou em relação à base
git log --oneline origin/{{repo.default_branch}}..origin/<branch>  # commits da branch (contexto que um PR daria)
git diff --stat origin/{{repo.default_branch}}...origin/<branch>   # resumo de arquivos/linhas alteradas
```

> Use a notação **three-dot** (`base...branch`) no `git diff` para revisar apenas as mudanças introduzidas pela branch, ignorando o que avançou na base em paralelo.

Confirme que o diff corresponde à issue. Se a branch estiver vazia em relação à base (sem diferenças), sinalize ao usuário antes de prosseguir.

---

## Passo 5 — Revisar o Diff

Revise o diff de forma estruturada, **ancorando cada observação nas Definições Técnicas** lidas no Passo 2. Cubra:

### 5.1 Aderência aos Critérios de Aceite Técnicos
Para cada critério de aceite técnico do grooming, verifique se o diff o satisfaz. Marque cada um como ✅ atende / ⚠️ parcial / ❌ não atende.

### 5.2 Corretude
- A lógica está correta? Há bugs, condições invertidas, off-by-one, promessas não aguardadas?
- Os **casos de borda** previstos no grooming foram tratados?
- O tratamento de erro é adequado (status HTTP corretos, mensagens, exceptions)?

### 5.3 Regras e Validações
- As **regras e validações técnicas** do grooming foram implementadas?
- Há validação de entrada (DTOs/validators no backend; checagens no frontend)?

### 5.4 Consistência com a Base
- O código segue os padrões e a stack já usados no repositório (backend em `{{repo.packages.backend}}`, frontend em `{{repo.packages.frontend}}`, tipos compartilhados em `{{repo.packages.shared}}` — quando existirem)?
- **Migrations:** se `repo.orm` não for `"none"`, mudanças de schema usam a ferramenta configurada ali (nunca MCP de banco de dados direto, tipo Supabase `apply_migration`/`execute_sql`). Se o projeto tiver um `CLAUDE.md` na raiz, siga também as convenções nele descritas.
- Não há segredos, tokens ou credenciais commitados.

### 5.5 Testes
- Há testes cobrindo o comportamento novo/alterado?
- Se `repo.coverage_config_path` estiver configurado, o backend mantém o threshold de cobertura lido ao vivo desse arquivo?

### 5.6 Segurança e Escopo
- Endpoints sensíveis estão protegidos pelo guard de auth?
- A branch está **dentro do escopo** da issue (sem mudanças não relacionadas)?

> Se durante a revisão você identificar um **bug ou pendência fora do escopo desta issue**, crie uma nova issue via skill `create-issue` (e, se houver um `CLAUDE.md` no projeto, siga também as convenções nele descritas) — não o conserte silenciosamente.

---

## Passo 6 — Escrever o Relatório de Revisão

Produza o relatório em **português brasileiro**, objetivo e acionável. Use o template abaixo.

```markdown
## Revisão de Código

**Issue:** #[número] [título]
**Branch:** [nome da branch]
**Revisor:** [usuário logado — `gh api user --jq .login`]
**Data:** [data atual]
**Veredito:** ✅ Aprovado | ❌ Reprovado

### Resumo
[2-4 linhas: o que a branch entrega e a conclusão geral da revisão.]

### Critérios de Aceite Técnicos
- [✅/⚠️/❌] [Critério] — [observação]
- ...

### Pontos Encontrados
Para cada ponto, indique severidade e local:
- **[Bloqueante | Importante | Sugestão]** `arquivo:linha` — [descrição do problema e correção esperada]
- ...

### Pontos Positivos
- [O que está bem implementado — reforço do que manter]

### Issues Criadas (se houver)
- [#número] — [bug/pendência fora do escopo identificado durante a revisão]

### Decisão
[Justificativa do veredito. Se reprovado, liste objetivamente o que precisa ser corrigido antes de uma nova revisão.]
```

**Regra de veredito:**
- **Aprovado** — não há pontos **Bloqueantes** e todos os critérios de aceite técnicos estão ✅ (ou ⚠️ justificados e aceitos).
- **Reprovado** — há ao menos um ponto **Bloqueante** ou um critério de aceite ❌.

---

## Passo 7 — Atualizar a Descrição da Issue

Obtenha o `body` atual com `gh issue view <n> --json body` e grave a nova descrição (ver "Anexar à descrição") contendo:
```
[descrição atual da issue]

---

[relatório de revisão produzido no Passo 6]
```

> Nunca sobrescreva a descrição da issue — o relatório de revisão é sempre **anexado ao final**, preservando o conteúdo anterior (Análise de Negócio + Definições Técnicas) intacto.

---

## Passo 8 — Mover o Status conforme o Veredito

Trocar o label de status de `{{statuses.in_review}}` para:
- **Aprovado** → `{{statuses.testing}}`.
- **Reprovado** → `{{statuses.in_develop}}`.

Confirmar ao usuário:
- Aprovado: *"Revisão concluída. A issue #[número] '[título]' foi aprovada e movida para Testing. Relatório adicionado à descrição."*
- Reprovado: *"Revisão concluída. A issue #[número] '[título]' foi reprovada e devolvida para In Develop. Os pontos a corrigir estão na descrição."*

---

## Princípios da Revisão

| Princípio | Aplicação |
|-----------|-----------|
| Baseado no grooming | Toda observação deve referenciar um critério de aceite ou regra técnica definida |
| Veredito objetivo | Aprovação/reprovação segue regra explícita — não "achismo" |
| Acionável | Cada ponto reprovante deve dizer o que corrigir, não só o que está errado |
| Preservar histórico | Revisão é anexada ao final da descrição; o conteúdo anterior nunca é sobrescrito |
| Bug fora do escopo vira issue | Pendências não relacionadas viram nova issue, não correção silenciosa |
| Escopo da branch | A branch deve fazer só o que a issue pede — sinalizar mudanças não relacionadas |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Mover a issue para In Review no início | A issue **já entra** em In Review — a primeira ação é ler o contexto |
| Usar `git diff base..branch` (two-dot) | Use **three-dot** (`base...branch`) para revisar só o que a branch introduziu |
| Esquecer o `git fetch` antes do diff | Sempre atualize as refs do `origin` antes de comparar |
| Revisar sem ler as Definições Técnicas | Sempre leia descrição + comentários antes de abrir o diff |
| Aprovar com ponto Bloqueante em aberto | Bloqueante = reprovação automática → volta para In Develop |
| Sobrescrever a descrição da issue | O relatório é sempre anexado ao final, nunca substitui o conteúdo anterior |
| Adivinhar qual é a branch | Se não houver vínculo claro, pergunte ao usuário |
| Corrigir bug fora do escopo na própria branch | Crie uma issue via skill `create-issue` e mantenha o foco da issue atual |
| Adicionar o novo label de status sem remover `in_review` | Sempre trocar o label em uma única chamada (`--remove-label` + `--add-label`) |
| Mover o status antes de atualizar a descrição | A descrição deve ser atualizada antes da mudança de status final |
| Executar sem `.opencode/dev-skills.config.json` | Rodar `project-setup` primeiro — nunca inventar repositório ou labels |
