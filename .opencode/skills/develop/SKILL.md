---
name: develop
description: Use when asked to implement, develop, or code a task — picks up a "ready to develop" GitHub issue, implements the code based on its Definições Técnicas, runs tests/lint/build, opens a PR and moves the issue to "in review". Triggers on phrases like "desenvolver a task", "implementar a issue", "codar o card", "fazer o develop", "pegar a próxima task pra desenvolver". Always invoke this skill for development work — never handle it manually.
---

# develop

## Overview

Conduz a implementação de uma issue do GitHub já refinada tecnicamente: lê as **Definições Técnicas** geradas pelo `grooming`, implementa o código no repositório com o **backend guiado por TDD** (quando o projeto separa backend/frontend), valida os testes do backend nos arquivos alterados com o gate de cobertura configurado no projeto + lint + build, abre um Pull Request e move a issue para o status de revisão.

Este fluxo fecha o pipeline do projeto:

```
create-issue → business-analysis → grooming → develop
   (o quê)        (o quê / por quê)  (como)    (implementação)
```

## Configuração do Projeto

Esta skill é agnóstica a projeto. Antes de qualquer chamada ao GitHub ou ao repositório, leia `.opencode/dev-skills.config.json` na raiz do projeto.

> **Se o arquivo não existir, ou não tiver a seção `github`:** informe o usuário — *"Este projeto ainda não foi configurado para as dev-skills. Rode a skill `project-setup` primeiro."* — e encerre sem executar nenhuma ação. Nunca invente repositório, labels de status ou caminhos de pacote.

| Referência nesta skill | Chave no config |
|---|---|
| Repositório das issues | `github.repo` |
| Labels de status | `statuses.*` |
| Remote do código | detectado via `git remote get-url origin` — nunca hardcode (pode ser diferente de `github.repo` se as issues vivem em um repositório central) |
| Branch base | `repo.default_branch` |
| Gerenciador de pacotes | `repo.package_manager` |
| Backend | `repo.packages.backend` (se ausente, projeto não separa backend — trate o pacote único como "backend" para fins de TDD/cobertura) |
| Frontend | `repo.packages.frontend` (se ausente, pule etapas de frontend) |
| Código compartilhado | `repo.packages.shared` (se ausente, pule) |
| Comando de teste | `repo.test_command` |
| Comando de lint | `repo.lint_command` |
| Comando de build | `repo.build_command` |
| Config de cobertura | `repo.coverage_config_path` (se ausente/`null`, projeto não tem gate de cobertura configurado — pule o gate) |
| ORM / ferramenta de migration | `repo.orm` (se `"none"`, projeto não usa migrations de banco — pule a seção de Banco de Dados) |

## Operações no GitHub

Todas as operações de issue usam a `gh` CLI, sempre com `-R {{github.repo}}`. Se o `gh` não estiver autenticado, oriente o usuário a rodar `gh auth login` e encerre.

| Operação | Comando |
|---|---|
| Ler issue | `gh issue view <n> -R {{github.repo}} --json number,title,body,labels,state,createdAt,url` |
| Ler comentários | `gh issue view <n> -R {{github.repo}} --json comments` |
| Listar por status | `gh issue list -R {{github.repo}} --state open --label "<label de status>" --json number,title,labels,createdAt --limit 100` |
| Mudar de status | `gh issue edit <n> -R {{github.repo}} --remove-label "<label de status atual>" --add-label "<novo label>"` |
| Anexar à descrição | escrever `body atual + "\n\n---\n\n" + conteúdo` em um arquivo temporário → `gh issue edit <n> -R {{github.repo}} --body-file <arquivo>` |

> **Label de status atual** = o label da issue que corresponde a algum valor de `statuses.*`. Se a issue tiver mais de um label de status (estado inconsistente), remova todos na troca e avise o usuário.

## Fluxo Obrigatório

```
[início]  → Verificar .opencode/dev-skills.config.json (se ausente: orientar rodar project-setup e encerrar)
         → Identificar issue (ready_to_develop) → Mover para in_develop
         → Ler todo o contexto (descrição + comentários)
         → Pré-flight git + criar branch a partir da branch base atualizada
         → Implementar via TDD no backend (red → green → refactor)
         → Migrations via ORM configurado (se houver mudança de schema)
         → Validar: testes do backend nos alterados (cobertura conforme configurado no repo) + lint + build
         → Self-review do diff
         → Commit(s)
         → Abrir PR via gh CLI
         → Vincular PR à issue → Mover para in_review
[fim]
```

**Nunca pule etapas. A ordem é obrigatória.**

---

## Passo 1 — Identificar a Issue

**Se um número ou URL de issue foi passado como argumento:** use-o diretamente via `gh issue view`.

**Se nenhuma issue foi especificada:**
1. Listar as issues abertas com o label `{{statuses.ready_to_develop}}` (ver "Listar por status").
2. Se houver múltiplas issues, listar número + título e perguntar ao usuário qual deseja desenvolver.
3. Se houver apenas uma, confirmar: *"Encontrei a issue #[número] '[título]'. Vou iniciar o desenvolvimento. Confirma?"*

**Se não houver issues em Ready To Develop:** informar o usuário e encerrar.

> Só desenvolva issues que passaram pelo grooming. Se a descrição **não** contém a seção `## Definições Técnicas`, avise o usuário que a issue precisa de grooming antes (skill `grooming`) e encerre.

---

## Passo 2 — Mover para In Develop

Trocar o label de status para `{{statuses.in_develop}}` (ver "Mudar de status") antes de escrever qualquer código.

---

## Passo 3 — Ler TODO o Contexto da Issue

A issue é a **fonte da verdade**. Tudo que as etapas anteriores produziram — análise de negócio (`business-analysis`) e definições técnicas (`grooming`) — existe para guiar e enriquecer esta implementação. O código entregue deve **satisfazer a regra de negócio e cumprir o objetivo da issue**, usando as Definições Técnicas como o *como*.

Executar em paralelo:
- `gh issue view <n> --json number,title,body,labels` — título, descrição completa (que já reúne a Análise de Negócio da `business-analysis` e as Definições Técnicas do `grooming`) e labels da issue
- `gh issue view <n> --json comments` — comentários adicionais (discussões e esclarecimentos que não fazem parte dos artefatos principais)

Leia o conjunto completo e extraia **duas camadas que devem ser honradas juntas**:

**1. Negócio (o quê / por quê) — da Análise de Negócio:**
- **Problema** e **objetivo** — o que a entrega precisa de fato resolver
- **Regras de Negócio** — restrições e políticas que o código deve fazer cumprir
- **Critérios de Aceitação** (Dado/Quando/Então) — o comportamento que valida a entrega
- **Escopo** — o que está **dentro** e, principalmente, o que está **fora** (não implemente além)
- **Riscos e Premissas** — condições que afetam decisões de implementação

**2. Técnico (como) — das Definições Técnicas:**
- **Entidades envolvidas** e mudanças de schema
- **Endpoints / serviços** a criar ou modificar
- **Mudanças no banco** e migrations necessárias
- **Regras e validações técnicas** e **casos de borda**
- **Critérios de Aceite Técnicos** — parte da checklist de "pronto"
- **Dependências e ordem de implementação**

> A implementação deve respeitar as **duas camadas simultaneamente**. As Definições Técnicas dizem *como* construir; as Regras de Negócio e os Critérios de Aceitação dizem *o que tem que ser verdade* ao final. Se houver conflito entre as camadas, **pare e pergunte ao usuário** — não escolha por conta própria.

Se houver **Perguntas Técnicas em Aberto** ou **Perguntas em Aberto** (de negócio) não resolvidas, pergunte ao usuário antes de codar — não invente decisões de arquitetura nem de regra de negócio. Respeite estritamente o **escopo**: não implemente o que foi marcado como fora.

---

## Passo 4 — Pré-flight Git e Criação da Branch

A branch é criada **automaticamente a partir da branch base atualizada (`repo.default_branch`), sempre antes de implementar qualquer linha de código**. Execute esta sequência na ordem:

**4.1 — Pré-flight: garantir working tree limpo.**
```bash
git status --porcelain
```
- Se a saída estiver **vazia**, prossiga.
- Se houver alterações não commitadas, **pare e pergunte ao usuário** o que fazer (commitar, fazer stash ou descartar). Nunca crie a branch nem implemente por cima de trabalho pendente.

**4.2 — Atualizar a branch base:**
```bash
git checkout {{repo.default_branch}} && git pull origin {{repo.default_branch}}
```

**4.3 — Criar a branch a partir da branch base e já mudar para ela:**
```bash
git checkout -b <nome-da-branch>
```

Convenção de nome:
- **Feature/Task:** `feat/<numero>-<slug-do-titulo>`
- **Bug:** `fix/<numero>-<slug-do-titulo>` (issue com o label `task_types.bug`)

Onde `<slug-do-titulo>` é o título em kebab-case e `<numero>` é o número da issue no GitHub. Ex.: `feat/42-refresh-token-cookie`.

> **A branch deve existir e estar ativa antes do Passo 5.** Nunca implemente, commite ou trabalhe direto na branch base. Confirme com `git branch --show-current` que está na branch nova antes de seguir.

---

## Passo 5 — Implementar (Guiado por TDD)

O desenvolvimento **do backend** é **guiado por testes (TDD)**. O teste vem **antes** do código de produção. Isso não é opcional: os **Critérios de Aceitação** (negócio) e os **casos de borda** (Definições Técnicas) já são a especificação dos testes — traduza-os em testes que falham e depois faça-os passar. (O frontend não passa por TDD/cobertura neste fluxo — ver abaixo. Se o projeto não separa backend/frontend, trate o pacote único como "backend" para efeito desta regra.)

Implemente seguindo a **ordem de implementação** das Definições Técnicas. Para cada unidade de comportamento, aplique o ciclo **Red → Green → Refactor**:

### Ciclo TDD (por critério/comportamento)

1. **🔴 Red — escreva um teste que falha.**
   - Traduza **um** Critério de Aceitação (Dado/Quando/Então) ou caso de borda em um teste.
   - Rode e **confirme que ele falha pelo motivo certo** (comportamento ausente, não erro de import/sintaxe). Um teste que nunca foi visto falhar não prova nada.

2. **🟢 Green — escreva o código mínimo para passar.**
   - Apenas o suficiente para o teste ficar verde. Sem antecipar requisitos não testados.

3. **🔵 Refactor — limpe com a rede de segurança verde.**
   - Melhore nomes, remova duplicação, alinhe ao padrão do repositório — mantendo todos os testes verdes.

Repita o ciclo até **todos os Critérios de Aceitação e casos de borda** terem teste correspondente e estarem verdes. Faça commits pequenos a cada ciclo coeso (ou a cada poucos ciclos), mantendo a suíte verde — alinhado aos commits atômicos do Passo 8.

> **Cobertura por especificação, não por métrica.** O alvo não é uma % de cobertura — é que cada Critério de Aceitação e cada caso de borda das Definições Técnicas tenha um teste que o prove. Se um critério é difícil de testar, isso costuma sinalizar um problema de design — ajuste o design, não pule o teste.

### Backend (`{{repo.packages.backend}}`)
- Testes unitários e de integração seguindo a convenção de testes já usada no repositório; fluxos ponta a ponta quando o critério for de integração.
- Estruture o código seguindo o padrão existente no pacote.

### Frontend (`{{repo.packages.frontend}}`, se o projeto tiver um)
- **Sem TDD e sem testes automatizados neste fluxo** — a cobertura de testes da skill é exclusiva do backend.
- Siga os padrões de componentes e a estrutura já existentes; valide o comportamento manualmente (ou via skill `verify`) quando aplicável.
- Para diretrizes de design visual ao criar UI nova, considere a skill `frontend-design`.

> **Escopo do TDD:** aplica-se ao **backend** (ou ao pacote único, se o projeto não separar backend/frontend). Toda **regra de negócio, validação, endpoint e caso de borda** do backend segue TDD sem exceção. Mudanças triviais de config no backend podem dispensar o teste-primeiro.

### Banco de Dados — **REGRA CRÍTICA**
Se `repo.orm` for `"none"`, pule esta seção — o projeto não usa migrations de banco.

Caso contrário, mudanças de schema vão **exclusivamente** pela ferramenta configurada em `repo.orm` (ex.: TypeORM, Prisma, Alembic, golang-migrate). **Nunca** use ferramentas MCP de banco de dados (Supabase ou equivalente) para criar ou alterar schema diretamente — elas são permitidas **apenas para inspeção read-only** (listar tabelas, ver logs/advisors).

1. Defina/atualize as entidades ou arquivos de schema seguindo a convenção de `repo.orm` no pacote de backend (`{{repo.packages.backend}}`).
2. Gere e rode a migration usando os comandos que o próprio `repo.orm` disponibiliza no projeto (ex.: `migration:generate`/`migration:run` no TypeORM, `alembic revision --autogenerate`/`alembic upgrade head` no Alembic). Se o config não tiver comandos explícitos para isso, confirme com o usuário antes de assumir a sintaxe.

### Bugs e pendências descobertos durante a implementação
Todo bug, débito técnico ou pendência fora do escopo desta issue **deve virar uma nova issue no GitHub** (não corrija silenciosamente fora de escopo). Use a skill `create-issue`, incluindo arquivo/linha, comportamento atual e esperado.

---

## Passo 6 — Validar (Testes do Backend nos Alterados + Cobertura + Lint + Build)

Os **testes automatizados rodam apenas no backend**, restritos aos **arquivos alterados** desta branch. Tudo deve passar antes de abrir o PR.

### 6.1 — Backend: testar apenas os arquivos alterados
```bash
cd {{repo.packages.backend}} && {{repo.test_command}}
```

> Rode `repo.test_command` restringindo aos arquivos afetados pelo diff desta branch quando o framework de testes do projeto suportar isso (flags de "changed since"/"related tests"); caso a ferramenta não suporte escopo por diff, rode a suíte do pacote normalmente. Nunca pule esta etapa. Se algum arquivo de produção alterado no backend **não tiver teste relacionado**, escreva o teste (volte ao TDD do Passo 5) — código de backend alterado sem teste não passa nesta etapa.

### 6.2 — Cobertura (backend, conforme configurado no repositório)
Se `repo.coverage_config_path` estiver definido, leia ali o threshold e o escopo de cobertura configurados (ex.: `coverageThreshold`/`collectCoverageFrom` em `jest.config.*`/`package.json`, `[tool.coverage]` em `pyproject.toml`, `.coveragerc`, ou o script de cobertura do projeto para outras stacks) — **não assuma um percentual ou caminho fixo**; use o que estiver configurado ali como fonte da verdade.

- Se `repo.coverage_config_path` for ausente/`null`, informe o usuário e siga sem esse gate — não invente um piso.
- Se a cobertura **reprovar** contra o threshold configurado, adicione testes (priorizando o que estiver descoberto — ramos `if/else`, `try/catch`, early returns, ternários e coalescência; funções e linhas não exercitadas) até o gate passar. Não prossiga com o gate vermelho.
- Se a issue tocar áreas fora do escopo de cobertura atual, considere ampliar essa configuração no arquivo real do projeto e **registre/alinhe** a mudança (não silenciosamente) — nunca duplique esse valor em outro lugar.

> Não baixe o threshold configurado só para fazer o gate passar — isso enfraquece o padrão do repositório. O caminho é escrever testes, não afrouxar a trava.

### 6.3 — Lint e Build (pacotes afetados)
Backend é obrigatório; o frontend entra apenas se foi alterado e o projeto tiver um pacote de frontend.
```bash
# backend (sempre que houver mudança no backend)
cd {{repo.packages.backend}} && {{repo.lint_command}} && {{repo.build_command}}
# frontend (apenas se o frontend foi alterado e existir, usando os mesmos comandos configurados)
cd {{repo.packages.frontend}} && {{repo.lint_command}} && {{repo.build_command}}
```

> Se algo falhar, **corrija antes de prosseguir**. Nunca abra o PR com testes do backend vermelhos, gate de cobertura reprovado, ou build quebrado. Se a falha for pré-existente e não relacionada à issue, registre uma nova issue (skill `create-issue`) e informe o usuário.

### Conferência dos Critérios de Aceitação
Antes de seguir, revise a implementação contra os critérios extraídos no Passo 3 — **ambas as camadas**:
- [ ] Cada **Critério de Aceitação de negócio** (Dado/Quando/Então) tem teste e é satisfeito pelo código
- [ ] Cada **Critério de Aceite Técnico** é verdadeiro
- [ ] Todas as **Regras de Negócio** estão de fato impostas pelo código (não só documentadas)
- [ ] Os **casos de borda** das Definições Técnicas estão tratados e testados
- [ ] **Gate de cobertura configurado no repositório** passando no backend
- [ ] Nada **fora do escopo** foi implementado

Se algum critério não puder ser cumprido com o contexto disponível, **pare e alinhe com o usuário** antes de abrir o PR.

---

## Passo 7 — Self-Review do Diff

Antes de commitar, faça uma **revisão crítica do próprio código** sobre o diff completo da branch. O objetivo é entregar em `In Review` um PR já limpo.

1. Gere o diff acumulado contra a branch base:
   ```bash
   git diff {{repo.default_branch}}...HEAD
   ```
   (e `git status` para conferir arquivos novos ainda não rastreados)

2. Rode a skill de revisão do repositório sobre o diff:
   - **`/code-review`** — busca bugs de correção no diff
   - **`/simplify`** — reuso, simplificação e limpeza (sem caçar bugs)

   Se as skills não estiverem disponíveis no contexto, faça a revisão manualmente cobrindo os mesmos eixos abaixo.

3. Revise contra esta checklist e **corrija o que encontrar** antes do commit:
   - [ ] Sem código morto, `console.log`/`print` de depuração, comentários `TODO` órfãos
   - [ ] Sem segredos, tokens ou credenciais hardcoded
   - [ ] Nomes, padrões e estilo coerentes com o código ao redor
   - [ ] Tratamento de erros e dos **casos de borda** das Definições Técnicas
   - [ ] Sem alterações fora do escopo da issue vazando no diff
   - [ ] Migrations geradas conferem com as mudanças de entidade

> Achados de correção que **não** pertencem ao escopo desta issue viram nova issue (skill `create-issue`), não correção silenciosa. Se o self-review exigir mudanças relevantes, **revalide** (Passo 6) antes de prosseguir.

---

## Passo 8 — Commit (Conventional Commits)

Faça **commits atômicos** (cada commit é uma unidade lógica coerente que, idealmente, mantém a suíte verde) seguindo o padrão [Conventional Commits](https://www.conventionalcommits.org). Mensagens em **português brasileiro**.

**Formato:**
```
<tipo>(<escopo>): <descrição no imperativo, minúscula, sem ponto final>

<corpo opcional: o quê e por quê — não o como. Quebre em ~72 colunas.>

<rodapé opcional: refs, breaking change, co-autoria>
```

**Tipos permitidos:**

| Tipo | Uso |
|------|-----|
| `feat` | Nova funcionalidade |
| `fix` | Correção de bug |
| `refactor` | Mudança que não altera comportamento externo |
| `perf` | Melhoria de performance |
| `test` | Adição/ajuste de testes |
| `docs` | Documentação |
| `style` | Formatação, sem mudança de lógica |
| `build` | Build, dependências, configs de empacotamento |
| `ci` | Pipelines e automação |
| `chore` | Tarefas de manutenção sem impacto em src |
| `revert` | Reversão de commit anterior |

**Regras:**
- **Escopo** = módulo/área afetada, ex.: `auth`, `users`, `frontend`, `shared` (adapte aos módulos reais do projeto).
- **Descrição** no imperativo ("adiciona", "corrige", não "adicionado"/"adicionando"), minúscula, **sem ponto final**, idealmente ≤ 50 e no máximo 72 caracteres.
- **Breaking change:** use `!` após o escopo (`feat(auth)!: ...`) e/ou rodapé `BREAKING CHANGE: <descrição>`.
- Referencie a issue no rodapé (`Refs: #<numero>`) e inclua a co-autoria.

**Exemplo:**
```
feat(auth): adiciona refresh de token via cookie httpOnly

Implementa POST /auth/refresh conforme as Definições Técnicas da issue.
O refresh token passa a ser armazenado em cookie httpOnly em vez de
localStorage, reduzindo a superfície de XSS.

Refs: #42
Co-Authored-By: Claude <noreply@anthropic.com>
```

> Faça commit/push apenas quando o usuário pedir desenvolvimento — o que é o caso aqui. Nunca commite na branch base.

---

## Passo 9 — Abrir o Pull Request

1. Push da branch:
   ```bash
   git push -u origin <nome-da-branch>
   ```
2. Abra o PR com a `gh` CLI, tendo como base a branch configurada em `repo.default_branch`:
   ```bash
   gh pr create --base {{repo.default_branch}} --title "<título>" --body-file <arquivo com o corpo>
   ```

**Título do PR:** mesmo padrão Conventional Commits da mensagem de commit — `<tipo>(<escopo>): <descrição>` — geralmente refletindo a entrega principal da issue. Ex.: `feat(auth): adiciona refresh de token via cookie httpOnly`.

> PRs **pequenos e focados** na issue. Se o PR ficou grande, é sinal de escopo misturado — extraia o que não pertence à issue para outra issue.

> **Use `Refs #<numero>`, nunca `Closes`/`Fixes`/`Resolves #<numero>`.** Palavras-chave de fechamento fechariam a issue no merge, pulando os estágios seguintes do pipeline (testing, homologating, deploy).

**Corpo do PR** — use o template abaixo:
```markdown
## Contexto
[1-2 linhas ligando ao problema/objetivo de negócio da issue]

## O que mudou
- [Mudança principal, mapeada aos Critérios de Aceite]
- [...]

## Como testar
1. [Passo a passo para validar localmente]

## Migrations
[Liste as migrations incluídas, ou "Nenhuma"]

## Checklist
- [ ] Testes, lint e build passando
- [ ] Critérios de Aceitação (negócio e técnicos) atendidos
- [ ] Sem mudanças fora do escopo
- [ ] Self-review realizado

## Issue
Refs #<numero>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

Se `github.repo` for diferente do repositório do código, use a referência completa (`Refs <owner>/<repo>#<numero>`). Capture a URL do PR retornada — ela vai na descrição da issue e no fechamento.

---

## Passo 10 — Vincular o PR e Mover para In Review

1. Adicionar o link do PR à descrição da issue (ver "Anexar à descrição"): descrição atual + `\n\n---\n\n` + `**PR aberto:** [URL]`.
2. Trocar o label de status de `{{statuses.in_develop}}` para `{{statuses.in_review}}`.

Confirmar ao usuário: *"Desenvolvimento concluído. PR aberto ([URL]) e a issue #[número] '[título]' foi movida para In Review."*

---

## Princípios do Desenvolvimento

| Princípio | Aplicação |
|-----------|-----------|
| Fiel à regra de negócio | A entrega deve cumprir o objetivo, as Regras de Negócio e os Critérios de Aceitação da issue — não só a parte técnica |
| Fiel ao grooming | Cada decisão de *como* implementar deve ser rastreável a uma Definição Técnica |
| Escopo é lei | Implemente o que está dentro do escopo; nunca o que foi marcado como fora |
| Conflito → perguntar | Divergência entre negócio e técnico, ou ambiguidade não resolvida, para a implementação e vira pergunta |
| Teste primeiro (TDD no backend) | Toda regra/validação/endpoint do backend nasce de um teste que falha (red → green → refactor) |
| Cobertura conforme o repositório | Gate lido ao vivo de `repo.coverage_config_path` — nunca afrouxar para passar |
| Testar só o backend e escopo do diff quando possível | Validação roda `repo.test_command` restrito aos arquivos alterados quando a ferramenta suportar, não a suíte de frontend |
| Padrão do repositório | Código novo deve se parecer com o código existente ao redor |
| Schema só via ORM configurado | Nunca alterar o banco pelas MCP tools de banco de dados — sempre pela ferramenta de `repo.orm` |
| Verde antes de PR | Testes, cobertura, lint e build devem passar antes de abrir o PR |
| Escopo disciplinado | Achou bug/pendência fora do escopo? Vira issue, não correção silenciosa |
| Status em tempo real | O label de status da issue sempre reflete o estado real do trabalho |

## Erros Comuns

| Erro | Correção |
|------|----------|
| Desenvolver issue sem grooming | Exigir a seção `## Definições Técnicas`; senão, encaminhar para a skill `grooming` |
| Implementar só o lado técnico | Conferir o código contra os Critérios de Aceitação e Regras de Negócio, não só os técnicos |
| Codar além do escopo | Respeitar o "fora do escopo" da análise de negócio; extra vira issue |
| Decidir sozinho em conflito negócio×técnico | Parar e perguntar ao usuário antes de codar |
| Escrever código de backend antes do teste | Seguir TDD: o teste que falha vem primeiro |
| Baixar o threshold de cobertura configurado para o gate passar | Manter o threshold do repositório; escrever testes nos pontos descobertos, nunca afrouxar |
| Testar o frontend ou ignorar o escopo do diff quando a ferramenta suporta | Usar `repo.test_command` do backend, restrito aos arquivos alterados quando possível |
| Alterar schema via MCP de banco de dados direto | Sempre usar entidades + migrations da ferramenta de `repo.orm` |
| Trabalhar/commitar na branch base | Sempre criar branch `feat/` ou `fix/` a partir da branch base |
| Abrir PR com a suíte vermelha | Rodar testes + cobertura + lint + build e corrigir antes |
| Corrigir bugs fora do escopo no mesmo PR | Registrar nova issue e manter o PR focado na issue atual |
| Usar `Closes #n` no PR | Usar `Refs #n` — o fechamento da issue acontece só no estágio de deploy |
| Mover para In Review antes do PR | O PR é aberto antes da mudança de status final |
| Esquecer migrations no PR | Listar migrations geradas no corpo do PR |
| Executar sem `.opencode/dev-skills.config.json` | Rodar `project-setup` primeiro — nunca inventar repositório, labels ou caminhos |

## Referências — Padrão de Git

| Fonte | O que define |
|-------|--------------|
| [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) | Estrutura `tipo(escopo): descrição`, tipos, `BREAKING CHANGE`, footers |
| [Angular Commit Message Guidelines](https://github.com/angular/angular/blob/main/contributing-docs/commit-message-guidelines.md) | Origem da convenção; conjunto de tipos e escopos |
| ["How to Write a Git Commit Message" — Chris Beams](https://cbeams.github.io/git-commit/) | As 7 regras: imperativo, linha de assunto ≤ 50, corpo a 72 colunas, o quê/por quê |
| [Semantic Versioning 2.0.0](https://semver.org/) | Relação `feat`/`fix`/`BREAKING CHANGE` → minor/patch/major |
| [Pro Git — Contributing to a Project](https://git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project) | Boas práticas de commits atômicos e mensagens |
