# URL Shortener — Infrastructure Setup Guide

### Инструкция по развёртыванию url_shortener на VirtualBox с k3s.

## Фаза 0 — Настройка сети VirtualBox

### Host-only сеть

VirtualBox → Tools → Networks → Host-only Networks:

- **vboxnet1**: IPv4 Prefix `192.168.57.1/24`, DHCP Server — Disabled

> Дефолтный vboxnet0 необходимо оставить
> 
> Если его нет то создать со следующими параметрами:
> 
> `name: vboxnet0, ipv4: 192.168.56.1/24, DHCP: enabled`

> VirtualBox может сбрасывать IP при перезапуске. Если это произошло, зафиксировать через CLI:
> ```bash
> VBoxManage hostonlyif ipconfig vboxnet1 --ip 192.168.57.1 --netmask 255.255.255.0
> ```

Проверка на хосте:
```bash
ip addr show vboxnet1
# Должен показать inet 192.168.57.1/24
```

---

## Фаза 1 — VM1 (k3s node)

### Создание VM

- RAM: 4GB (минимум 2GB)
- Disk: 20GB
- Adapters:
  - Adapter 1: NAT
  - Adapter 2: Host-only Adapter → vboxnet1

### Установка Ubuntu Server 22.04

При установке:
- Network:
  - enp0s3 — DHCP (оставить по умолчанию),
  - enp0s8 — static `192.168.57.20/24` (Enter -> Edit IPV4 -> Manual -> Subnet: `192.168.57.0/24`, Address: `192.168.57.20`)
- Отметить "Install OpenSSH Server"
- Остальное можно оставить как есть

Пример:


![network](./network.png)

### Проверка после установки

```bash
# С хоста
ssh user@192.168.57.20
# На VM
ping google.com
```

---

## Фаза 2 — VM2 (databases)

### Создание VM

- RAM: 2GB
- Disk: 20GB
- Adapters: аналогично VM1

### Установка Ubuntu Server 22.04
Аналогично VM1
- Network:
  - enp0s3 — DHCP (оставить по умолчанию),
  - enp0s8 — static `192.168.57.21/24` (Enter -> Edit IPV4 -> Manual -> Subnet: `192.168.57.0/24`, Address: `192.168.57.21/24`)
- Отметить "Install OpenSSH Server"
- Остальное можно оставить как есть

### PostgreSQL 16

```bash
sudo apt update
sudo apt install -y gnupg2
sudo sh -c 'echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list'
curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | sudo gpg --dearmor -o /etc/apt/trusted.gpg.d/postgresql.gpg
sudo apt update
sudo apt install -y postgresql-16
```

Создание баз и пользователей:

```bash
sudo -u postgres psql <<'SQL'
CREATE USER urlshortener WITH PASSWORD 'YOUR_URLS_PASSWORD';
CREATE USER iam_user WITH PASSWORD 'YOUR_IAM_PASSWORD';

CREATE DATABASE urlshortener_db OWNER urlshortener;
CREATE DATABASE iam_db OWNER iam_user;

REVOKE ALL ON DATABASE urlshortener_db FROM PUBLIC;
REVOKE ALL ON DATABASE iam_db FROM PUBLIC;
SQL
```

Открыть для сети:

```bash
# Слушать на всех интерфейсах
sudo sed -i "s/#listen_addresses = 'localhost'/listen_addresses = '*'/" /etc/postgresql/16/main/postgresql.conf

# Разрешить подключения с VM1
echo '# Allow VM1 (k3s)
host    urlshortener_db    urlshortener    192.168.57.20/32    scram-sha-256
host    iam_db             iam_user        192.168.57.20/32    scram-sha-256' | sudo tee -a /etc/postgresql/16/main/pg_hba.conf

sudo systemctl restart postgresql
```

### Redis 7

```bash
sudo apt install -y redis-server
```

Открыть для сети:

```bash
sudo sed -i 's/^bind .*/bind 127.0.0.1 192.168.57.21/' /etc/redis/redis.conf
sudo systemctl restart redis-server
```

> Один инстанс Redis, два сервиса используют разные DB номера:
> - URLSService → DB 0
> - IAM Service → DB 1

### Firewall (UFW)

```bash
sudo ufw allow from 192.168.57.20 to any port 5432   # PostgreSQL
sudo ufw allow from 192.168.57.20 to any port 6379   # Redis
sudo ufw allow ssh
sudo ufw enable
```

### Проверка с VM1

```bash
psql -h 192.168.57.21 -U urlshortener -d urlshortener_db
psql -h 192.168.57.21 -U iam_user -d iam_db
redis-cli -h 192.168.57.21 -p 6379 ping   # → PONG
```

---

## Фаза 3 — k3s на VM1

### Установка

```bash
# На VM1
curl -sfL `https://get.k3s.io` | sh -

# Проверить
sudo kubectl get nodes
```
> Если не можем достучаться, то копируем c файл k3s_install_script.sh на vm1 через scp
> ```bash
> # С хоста
> scp /path/to/script/dir/k3s_install_script.sh <user_name>@192.168.57.20:/home/<user_name>/
> 
> # На VM1
> cd ~/
> sudo chmod +x k3s_install_script.sh; sudo ./k3s_install_script.sh
> ```
 

### TLS SAN для доступа с хоста

```bash
echo 'tls-san: "192.168.57.20"' | sudo tee /etc/rancher/k3s/config.yaml

# Пересоздать сертификаты
sudo systemctl stop k3s
sudo rm /var/lib/rancher/k3s/server/tls/dynamic-cert.json
sudo rm /var/lib/rancher/k3s/server/tls/serving-kube-apiserver.crt
sudo rm /var/lib/rancher/k3s/server/tls/serving-kube-apiserver.key
sudo systemctl start k3s
```

### kubectl с хост-машины

На VM1 — скопировать kubeconfig:

```bash
sudo cp /etc/rancher/k3s/k3s.yaml /tmp/k3s.yaml
sudo chown $USER:$USER /tmp/k3s.yaml
```

На хосте:

```bash
mkdir -p ~/.kube
scp user@192.168.57.20:/tmp/k3s.yaml ~/.kube/config-k3s
sed -i 's|127.0.0.1|192.168.57.20|' ~/.kube/config-k3s

# Сделать дефолтным
echo 'export KUBECONFIG=~/.kube/config-k3s' >> ~/.bashrc
export KUBECONFIG=~/.kube/config-k3s

# Проверить
kubectl get nodes
```

---

## Фаза 4 — Сборка и загрузка Docker-образов

### Сборка образов (на хосте)

```bash
cd /path/to/url_shortener

# Frontend (production)
docker build -f URLSFrontend/Dockerfile --target runtime -t urls-frontend:latest URLSFrontend/

# URLs Service
docker build -f URLSService/docker/Dockerfile -t urls-service:latest URLSService/

# IAM Service
docker build -f urls_iam_service/Dockerfile -t iam-service:latest urls_iam_service/

# Migration images
docker build -f k8s/Dockerfile.migrations --build-arg MIGRATIONS_PATH=db/migrations -t urls-migration:latest URLSService/
docker build -f k8s/Dockerfile.migrations --build-arg MIGRATIONS_PATH=migrations -t iam-migration:latest urls_iam_service/
```

### Загрузка в k3s

```bash
# Сохранить образы
docker save urls-frontend:latest -o /tmp/urls-frontend.tar
docker save urls-service:latest -o /tmp/urls-service.tar
docker save iam-service:latest -o /tmp/iam-service.tar
docker save urls-migration:latest -o /tmp/urls-migration.tar
docker save iam-migration:latest -o /tmp/iam-migration.tar

# Скопировать на VM1
scp /tmp/urls-frontend.tar /tmp/urls-service.tar /tmp/iam-service.tar /tmp/urls-migration.tar /tmp/iam-migration.tar user@192.168.57.20:/tmp/

# Импортировать в k3s (на VM1)
ssh user@192.168.57.20 'sudo k3s ctr images import /tmp/urls-frontend.tar && sudo k3s ctr images import /tmp/urls-service.tar && sudo k3s ctr images import /tmp/iam-service.tar && sudo k3s ctr images import /tmp/urls-migration.tar && sudo k3s ctr images import /tmp/iam-migration.tar'
```

---

## Фаза 5 — Деплой в k3s

### Структура манифестов

```
k8s/
├── namespace.yaml          # Namespace urls
├── secrets.yaml            # Пароли к БД и Redis
├── frontend.yaml           # Frontend Deployment + Service
├── urls-service.yaml       # URLSService Deployment + Service (с init container для миграций)
├── iam-service.yaml        # IAM Service Deployment + Service (с init container для миграций)
├── ingress.yaml            # Traefik IngressRoute + ForwardAuth Middleware (маршрутизация + авторизация)
└── Dockerfile.migrations   # Dockerfile для сборки migration-образов
```

### Применение

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/frontend.yaml
kubectl apply -f k8s/urls-service.yaml
kubectl apply -f k8s/iam-service.yaml
kubectl apply -f k8s/ingress.yaml
```

### Проверка

```bash
# Статус подов
kubectl get pods -n urls

# Логи миграций
kubectl logs -n urls -l app=urls-service -c migrate
kubectl logs -n urls -l app=iam-service -c migrate

# Логи сервисов
kubectl logs -n urls -l app=urls-service
kubectl logs -n urls -l app=iam-service

# Ingress
kubectl get ingress -n urls
```

### Маршрутизация (Traefik IngressRoute)

Используются Traefik CRD вместо стандартного Ingress — для поддержки ForwardAuth middleware (авторизация через IAM-сервис, как в Docker-конфигурации).

`ingress.yaml` содержит:
- **Middleware `auth-required`** — ForwardAuth, вызывает `iam-service:9092/auth/validate`, пробрасывает заголовки `X-User-Id` и `X-Provider`
- **IngressRoute** — маршруты с приоритетами

| Приоритет | Match | Middleware | Сервис |
|-----------|-------|------------|--------|
| 20 | `/api/v1/health`, `/api/v1/readiness` | — | urls-service:9091 |
| 15 | `/auth/*`, `/users/*` | — | iam-service:9092 |
| 12 | `GET /api/v1/urls` | auth-required | urls-service:9091 |
| 11 | `POST/DELETE/PUT/PATCH /api/*` | auth-required | urls-service:9091 |
| 10 | `GET /api/*` | — | urls-service:9091 |
| 1 | `/*` | — | urls-frontend:80 |

Приложение доступно по адресу: `http://192.168.57.20`

---

## Обновление сервисов (на данном этапе)

При изменении кода:

```bash
# 1. Пересобрать нужный образ
docker build -f URLSService/docker/Dockerfile -t urls-service:latest URLSService/

# 2. Загрузить в k3s
docker save urls-service:latest -o /tmp/urls-service.tar
scp /tmp/urls-service.tar user@192.168.57.20:/tmp/
ssh user@192.168.57.20 'sudo k3s ctr images import /tmp/urls-service.tar'

# 3. Перезапустить деплоймент
kubectl rollout restart deployment urls-service -n urls
```

При добавлении миграций — также пересобрать и загрузить migration-образ.

---

## Фаза 6 — Мониторинг (Prometheus + Grafana)

### Helm на хосте

```bash
curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### kube-prometheus-stack

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --timeout 10m
helm repo update --timeout 10m

helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  -f k8s/monitoring-values.yaml \
  --timeout 10m
```

Grafana: `http://192.168.57.20:30300` (логин `admin` / `prom-operator`)
Prometheus: `http://192.168.57.20:30090`

### Экспортеры на VM2

```bash
# Установка
sudo apt install -y prometheus-node-exporter prometheus-postgres-exporter prometheus-redis-exporter

# Настройка postgres_exporter
echo 'DATA_SOURCE_NAME="postgresql://urlshortener:YOUR_URLS_PASSWORD@localhost:5432/urlshortener_db?sslmode=disable"' | sudo tee /etc/default/prometheus-postgres-exporter
sudo systemctl restart prometheus-postgres-exporter

# Проверка (порты 9100, 9187, 9121)
ss -tlnp | grep -E '9100|9187|9121'
```

Firewall на VM2:

```bash
sudo ufw allow from 192.168.57.20 to any port 9100   # node_exporter
sudo ufw allow from 192.168.57.20 to any port 9187   # postgres_exporter
sudo ufw allow from 192.168.57.20 to any port 9121   # redis_exporter
```

Scrape-конфиг уже включён в `k8s/monitoring-values.yaml`.

### Дашборды Grafana

Импорт: Dashboards → New → Import, ввести ID:

| ID | Дашборд | Описание |
|----|---------|----------|
| 1860 | Node Exporter Full | CPU, RAM, диск, сеть |
| 9628 | PostgreSQL Database | Подключения, запросы, размер БД |
| 11835 | Redis Dashboard | Память, команды, ключи |

При импорте выбрать Prometheus как data source.

---

## Фаза 6.5 — Traefik Dashboard

k3s использует встроенный Traefik. Для включения дашборда:

```bash
kubectl apply -f k8s/traefik-dashboard.yaml
```

Манифест содержит HelmChartConfig (включает `--api.insecure=true`) и NodePort-сервис (порт 30888). Ресурсы в `kube-system` — не управляются ArgoCD.

Traefik Dashboard: `http://192.168.57.20:30888/dashboard/` (слеш в конце обязателен)

---

## Фаза 7 — CI/CD (GitHub Actions + ArgoCD)

### GitHub Actions

Workflow файл: `.github/workflows/build-and-push.yaml`

При push в `master`:
1. `detect-changes` — определяет какие сервисы изменились (по путям файлов)
2. Для каждого изменённого сервиса — собирает Docker-образ и пушит в ghcr.io
3. Образы получают теги: `:latest` и `:SHA-коммита`

Для ручного запуска (все образы) добавлен триггер `workflow_dispatch`:
GitHub → Actions → Build and Push Images → Run workflow.

Registry: `ghcr.io/artemborodinevgenyevich/`
> При форке заменить все упоминания registry, user, owner на свои

### ArgoCD

Установка:

```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

> На VM1 может потребоваться импорт образа Redis если `public.ecr.aws` недоступен:
> ```bash
> sudo k3s ctr images pull docker.io/library/redis:8.2.3-alpine
> sudo k3s ctr images tag docker.io/library/redis:8.2.3-alpine public.ecr.aws/docker/library/redis:8.2.3-alpine
> ```

Открыть UI:

```bash
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "NodePort", "ports": [{"port": 443, "nodePort": 30443}]}}'
```

Получить пароль:

```bash
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

ArgoCD UI: `https://192.168.57.20:30443` (логин `admin`)

Если repo-server падает с liveness probe timeout (VirtualBox NAT медленно проксирует HTTPS, git-операции к GitHub могут занимать 15+ секунд):

```bash
# Увеличить таймауты probe и git-запросов
kubectl patch deployment argocd-repo-server -n argocd --type=json -p='[
  {"op": "replace", "path": "/spec/template/spec/containers/0/livenessProbe/timeoutSeconds", "value": 30},
  {"op": "replace", "path": "/spec/template/spec/containers/0/livenessProbe/periodSeconds", "value": 60},
  {"op": "replace", "path": "/spec/template/spec/containers/0/livenessProbe/failureThreshold", "value": 5},
  {"op": "replace", "path": "/spec/template/spec/containers/0/readinessProbe/timeoutSeconds", "value": 30},
  {"op": "replace", "path": "/spec/template/spec/containers/0/readinessProbe/periodSeconds", "value": 30}
]'

kubectl patch configmap argocd-cmd-params-cm -n argocd --type merge -p '{"data":{"reposerver.git.request.timeout":"60"}}'
```

### ArgoCD Application

Применяется вручную (один раз):

```bash
kubectl apply -f k8s/argocd-app.yaml
```

ArgoCD следит за папкой `k8s/` в репозитории. Используется Kustomize — `k8s/kustomization.yaml` перечисляет ресурсы и управляет image-тегами.

### ArgoCD Image Updater

Установка:

```bash
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj-labs/argocd-image-updater/stable/config/install.yaml
```

Конфигурация через CRD:

```bash
kubectl apply -f k8s/image-updater.yaml
```

Image Updater каждые 2 минуты проверяет ghcr.io на наличие новых образов и автоматически обновляет деплойменты.

### Полный цикл деплоя

```
git push → GitHub Actions → сборка образов → push в ghcr.io
                                                    ↓
         ArgoCD ← следит за k8s/ в Git ← синхронизирует манифесты
                                                    ↓
         Image Updater ← следит за ghcr.io ← обновляет деплойменты
```

### Доступы к UI

| Сервис | URL | Логин |
|--------|-----|-------|
| Приложение | `http://192.168.57.20` | — |
| Traefik Dashboard | `http://192.168.57.20:30888/dashboard/` | — |
| Grafana | `http://192.168.57.20:30300` | admin / (из secret) |
| Prometheus | `http://192.168.57.20:30090` | — |
| ArgoCD | `https://192.168.57.20:30443` | admin / (из secret) |

---

## Полезные команды

```bash
# Статус всех ресурсов
kubectl get all -n urls

# Описание пода (если не стартует)
kubectl describe pod <pod-name> -n urls

# Логи init container (миграции)
kubectl logs <pod-name> -n urls -c migrate

# Зайти в под
kubectl exec -it <pod-name> -n urls -- sh

# Удалить и пересоздать
kubectl delete -f k8s/urls-service.yaml
kubectl apply -f k8s/urls-service.yaml
```
