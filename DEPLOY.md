# Deploy (CasaOS + GHCR + Watchtower)

Fluxo: push na `main` → GitHub Actions gera a imagem arm64 → publica em
`ghcr.io/vinirb15/go-backup` (`latest` + SHA curto) → o Watchtower no servidor
detecta em até 120 s e recria o container.

## 1. Primeiro build e visibilidade do pacote no GHCR

1. Faça push na `main` (ou rode **Actions → Deploy → Run workflow**).
2. Quando terminar, abra **github.com/vinirb15 → Packages → go-backup → Package settings**.
3. Em **Danger Zone → Change visibility**, deixe **Public**. O primeiro push pode
   criar o pacote como privado mesmo com o repositório público. Com o pacote
   público, o servidor não precisa de login.
4. Em **Manage Actions access**, confirme que o repositório `go-backup` aparece
   com papel **Write**. Normalmente isso é automático, pela label
   `org.opencontainers.image.source`.

## 2. Login no ghcr.io (só se o pacote for privado)

Crie um **Personal access token (classic)** só com o escopo `read:packages`.
Para pull não precisa de `write:packages` nem `repo`. No servidor, como root:

```sh
echo '<TOKEN>' | sudo docker login ghcr.io -u vinirb15 --password-stdin
```

Isso grava `/root/.docker/config.json`. Descomente a linha do `config.json` em
`docker-compose.watchtower.yml`.

## 3. Instalar no CasaOS

1. Prepare a pasta de dumps. O container roda como UID 1000:
   ```sh
   sudo mkdir -p /DATA/AppData/db-backup/dump
   sudo chown -R 1000:1000 /DATA/AppData/db-backup
   ```
   Se for reaproveitar a pasta de dumps do app antigo, aplique o mesmo `chown`
   nela. Agora o caminho **dentro** do container é `/app/dump`, e não mais
   `/root/dump`.
2. Remova o app antigo (que usava a imagem buildada localmente).
3. **App Store → Custom Install → Import** e cole o `docker-compose.yml`.
   Preencha os placeholders `<...>` na tela antes de instalar.
4. Repita o import com o `docker-compose.watchtower.yml`.

## 4. Testar o primeiro deploy

```sh
docker logs db-backup             # deve mostrar "Agendador iniciado..."
docker logs -f watchtower         # acompanha as checagens/atualizações
```

Faça um commit qualquer na `main` e espere o workflow terminar. Em até ~2 min, o
log do Watchtower deve mostrar `Found new ... image` e a recriação do
`db-backup`. Confira com:

```sh
docker inspect db-backup --format '{{.Image}} {{.Config.Image}}'
```

Para forçar uma checagem imediata, sem esperar o intervalo:

```sh
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  nickfedor/watchtower:1 --run-once --label-enable --cleanup
```

## 5. Rollback

Cada build também publica a tag com o SHA curto do commit (ex.: `a1b2c3d`). As
tags ficam listadas na página do pacote no GitHub.

- **Rápido:** no CasaOS, edite o app `db-backup` e troque a imagem para
  `ghcr.io/vinirb15/go-backup:<sha>`. Como essa tag nunca muda, o Watchtower
  deixa o container nela. Volte para `:latest` quando quiser retomar as
  atualizações automáticas.
- **Definitivo:** `git revert <commit>` + push na `main`. O pipeline publica um
  novo `latest`, e o Watchtower aplica normalmente.
