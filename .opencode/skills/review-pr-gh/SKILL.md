---
name: review-pr-gh
description: Use when asked to review a pull request, do code review of a task in review, or process a GitHub issue that is "in review". Triggers on phrases like "revisar o PR", "fazer code review", "review da task", "revisar a task que está in review", "processar revisão", "revisão de PR", or any request to review the code of a developed task and decide if it advances to testing. Always invoke this skill for PR review work — never handle it manually.
---

# review-pr-gh

## Overview

Conduz a revisão de código de uma issue do GitHub que está em revisão (com PR aberto): lê as **Definições Técnicas** geradas no grooming, localiza o Pull Request correspondente no GitHub, revisa o diff contra os critérios de aceite técnicos, anexa um relatório de revisão à descrição da issue e move o status conforme o veredito — **Aprovado → Testing** ou **Reprovado → In Develop**.

Este fluxo é a etapa seguinte ao desenvolvimento no pipeline. Onde o `grooming` define *como* implementar, esta skill verifica se *o que foi implementado* atende ao que foi definido.

## Configuração do Projeto

Esta skill é agnóstica a projeto. Antes de qualquer chamada ao GitHub, leia `.opencode/dev-skills.config.json` na raiz do projeto.

> **Se o arquivo não existir, ou não tiver a seção `github`:** informe o usuário — *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* — e encerre sem executar nenhuma ação. Nunca invente repositório, labels de status ou caminhos de pacote.

| Referência nesta skill | Chave no config |
|---|---|
| Repositório das issues | `github.repo` |
| Labels de status | `statuses.*` |
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
         → Localizar o PR no GitHub
         → Ler o PR (gh pr view + gh pr diff)
         → Revisar o diff contra os critérios de aceite técnicos
         → Escrever o Relatório de Revisão → Adicionar à descrição da issue
         → Veredito: Aprovado → testing | Reprovado → in_develop
[fim]
```

**Nunca pule etapas. A ordem é obrigatória.**

> Diferente das outras skills, a issue **já entra** nesta skill no status "In Review" — portanto **não há mudança de status no início**. A primeira ação é ler o contexto da issue.

---

## Passo 1 — Identificar a Issue

**Se um número/URL de issue ou um número de PR foi passado como argumento:** use-o diretamente.
- Número/URL de issue → `gh issue view`.
- Número de PR → `gh pr view <número>` e, em seguida, localize a issue vinculada pelo `Refs #<n>` no corpo do PR (ver Passo 3).

**Se nada foi especificado:**
1. Listar as issues abertas com o label `{{statuses.in_review}}` (ver "Listar por status").
2. Se houver múltiplas issues, listar número + título e perguntar ao usuário qual deseja revisar.
3. Se houver apenas uma, confirmar: *"Encontrei a issue #[número] '[título]' em In Review. Vou iniciar a revisão do PR. Confirma?"*

**Se não houver issues em In Review:** informar o usuário e encerrar.

---

## Passo 2 — Ler Contexto da Issue

Executar em paralelo:
- `gh issue view <n> --json number,title,body,labels` — para obter título, descrição completa (que já reúne a **Análise de Negócio**, as **Definições Técnicas** do grooming e o link do PR adicionado pela `develop`) e labels da issue.
- `gh issue view <n> --json comments` — para obter comentários adicionais (discussões que não fazem parte dos artefatos principais).

Com o conteúdo coletado, extraia mentalmente:
- Os **Critérios de Aceite Técnicos** definidos no grooming (a referência principal da revisão).
- As **Entidades**, **Endpoints/Serviços** e **Mudanças no Banco** previstos.
- As **Regras e Validações Técnicas** e os **Casos de Borda** que o código deveria tratar.
- O **objetivo de negócio** original (para checar se a implementação não desviou do propósito).

---

## Passo 3 — Localizar o PR no GitHub

1. Procure a referência do PR, nesta ordem:
   - número/URL do PR na **descrição** da issue (linha `**PR aberto:**` adicionada pela skill `develop`), com os **comentários** como fallback;
   - PRs que referenciam a issue: `gh pr list --state open --search "<número> in:body"` (o `develop` escreve `Refs #<número>` no corpo do PR);
   - se nada for encontrado, liste os PRs abertos com `gh pr list` e correlacione pela branch (`feat/<número>-...` ou `fix/<número>-...`).
2. Se ainda assim não for possível vincular com confiança, **pergunte ao usuário** qual é o número/URL do PR. Nunca adivinhe.

> A skill depende do `gh` autenticado e de um remote no GitHub. Se `gh` falhar (não autenticado ou sem remote), informe o usuário e encerre sem mover o status.

---

## Passo 4 — Ler o Pull Request

Executar em paralelo:
- `gh pr view <número> --json title,body,state,author,baseRefName,headRefName,files,additions,deletions,reviewDecision` — metadados do PR.
- `gh pr diff <número>` — o diff completo a ser revisado.

Confirme que o PR está aberto e que o diff corresponde à issue. Se o PR já estiver `MERGED` ou `CLOSED`, sinalize ao usuário antes de prosseguir.

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
- O PR está **dentro do escopo** da issue (sem mudanças não relacionadas)?

> Se durante a revisão você identificar um **bug ou pendência fora do escopo desta issue**, crie uma nova issue via skill `create-issue` (e, se houver um `CLAUDE.md` no projeto, siga também as convenções nele descritas) — não o conserte silenciosamente.

---

## Passo 6 — Escrever o Relatório de Revisão

Produza o relatório em **português brasileiro**, objetivo e acionável. Use o template abaixo.

```markdown
## Revisão de Código

**Issue:** #[número] [título]
**PR:** [#número — título do PR]
**Revisor:** [usuário logado — `gh api user --jq .login`]
**Data:** [data atual]
**Veredito:** ✅ Aprovado | ❌ Reprovado

### Resumo
[2-4 linhas: o que o PR entrega e a conclusão geral da revisão.]

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

> Nunca sobrescreva a descrição da issue — o relatório de revisão é sempre **anexado ao final**, preservando o conteúdo anterior (Análise de Negócio + Definições Técnicas + PR) intacto.

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
| Escopo do PR | O PR deve fazer só o que a issue pede — sinalizar mudanças não relacionadas |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Mover a issue para In Review no início | A issue **já entra** em In Review — a primeira ação é ler o contexto |
| Revisar sem ler as Definições Técnicas | Sempre leia descrição + comentários antes de abrir o diff |
| Aprovar com ponto Bloqueante em aberto | Bloqueante = reprovação automática → volta para In Develop |
| Sobrescrever a descrição da issue | O relatório é sempre anexado ao final, nunca substitui o conteúdo anterior |
| Adivinhar qual é o PR | Se não houver vínculo claro, pergunte ao usuário |
| Corrigir bug fora do escopo no próprio PR | Crie uma issue via skill `create-issue` e mantenha o foco da issue atual |
| Adicionar o novo label de status sem remover `in_review` | Sempre trocar o label em uma única chamada (`--remove-label` + `--add-label`) |
| Mover o status antes de atualizar a descrição | A descrição deve ser atualizada antes da mudança de status final |
| Executar sem `.opencode/dev-skills.config.json` | Rodar `project-setup` primeiro — nunca inventar repositório ou labels |
