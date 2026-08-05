# Homeshare — Self-hosted файловый обменник

## Архитектура

Homeshare — это легковесный Go-сервис для безопасного обмена файлами внутри доверенного круга пользователей.

### Ключевые компоненты:

1. **HTTP Server** — слушает 127.0.0.1:8090, за ним стоит Caddy с TLS
2. **SQLite Database** — хранение метаданных, пользователей, сессий, логов
3. **Chunked Upload** — собственный протокол загрузки по чанкам (32 MB)
4. **Sharded Storage** — файлы хранятся в `/srv/media/fileshare/data/ab/cd/<fileid>`
5. **Rate Limiter** — глобальные лимиты скорости для внешних клиентов
6. **Traffic Counter** — учёт трафика по пользователям с grace logic
7. **Audit Log** — логирование всех значимых событий
8. **Security Log** — raw IP для fail2ban

### Поток данных:

```
Client → Caddy (TLS termination) → Homeshare (127.0.0.1:8090)
                                      ↓
                              SQLite + File Storage
```

## Структура проекта

```
homeshare/
├── cmd/homeshare/          # Точка входа, CLI команды
│   └── main.go
├── internal/
│   ├── config/             # Конфигурация YAML
│   ├── db/                 # SQLite подключения, миграции
│   ├── models/             # Модели данных
│   ├── auth/               # Пароли, TOTP, сессии
│   ├── admin/              # Админские handlers
│   ├── api/                # REST API handlers
│   ├── web/                # Web UI handlers, templates
│   ├── storage/            # Файловое хранилище
│   ├── upload/             # Chunked upload логика
│   ├── download/           # Download/streaming логика
│   ├── ratelimit/          # Rate limiting
│   ├── traffic/            # Traffic counters
│   ├── cleanup/            # Фоновая очистка
│   ├── audit/              # Audit logging
│   ├── securitylog/        # Security log для fail2ban
│   ├── settings/           # Runtime настройки
│   └── utils/              # Утилиты
├── web/
│   ├── templates/          # HTML шаблоны
│   └── static/             # CSS, JS, изображения
├── go.mod
├── go.sum
├── config.yaml             # Пример конфигурации
└── README.md
```

## SQLite Schema

См. `internal/db/schema.sql` — полная схема БД со всеми таблицами.

## API Endpoints

### Публичные (без авторизации)
- `GET /` — главная страница
- `POST /api/invite/activate` — активация invite кода

### Пользовательские (требуется сессия)
- `GET /dashboard` — дашборд пользователя
- `POST /api/uploads` — создать загрузку
- `HEAD /api/uploads/{id}` — статус загрузки
- `PATCH /api/uploads/{id}?offset=N` — загрузить chunk
- `POST /api/uploads/{id}/complete` — завершить загрузку
- `DELETE /api/uploads/{id}` — отменить загрузку
- `GET /api/files` — список файлов
- `GET /api/files/{id}` — скачать файл
- `DELETE /api/files/{id}` — удалить свой файл
- `POST /api/zip` — создать ZIP из нескольких файлов

### Админские (требуется admin сессия + TOTP)
- `GET /admin` — админ дашборд
- `GET /admin/people` — управление пользователями
- `POST /admin/people` — создать Person
- `PUT /admin/people/{id}` — обновить Person
- `DELETE /admin/people/{id}` — удалить Person
- `GET /admin/invites` — управление invite кодами
- `POST /admin/invites` — создать invite
- `GET /admin/sessions` — сессии пользователей
- `POST /admin/sessions/{id}/revoke` — отозвать сессию
- `GET /admin/files` — все файлы
- `PUT /admin/files/{id}` — изменить файл (quarantine, protected, etc)
- `DELETE /admin/files/{id}` — удалить файл
- `GET /admin/uploads` — активные загрузки
- `POST /admin/uploads/{id}/cancel` — отменить загрузку
- `GET /admin/quarantine` — карантин
- `POST /admin/quarantine/{id}/approve` — одобрить файл
- `GET /admin/traffic` — счётчики трафика
- `POST /admin/traffic/reset` — сбросить счётчики
- `GET /admin/audit` — audit log
- `GET /admin/settings` — настройки
- `PUT /admin/settings` — обновить настройки
- `GET /admin/ratelimits` — активные rate limit locks
- `POST /admin/ratelimits/{id}/reset` — сбросить lock

## Пример config.yaml

```yaml
listen: "127.0.0.1:8090"
base_url: "https://files.example.duckdns.org"

data_dir: "/srv/media/fileshare/data"
tmp_dir: "/srv/media/fileshare/tmp"
db_path: "/srv/media/fileshare/db/homeshare.db"
backup_dir: "/home/fileshare-backup"

local_cidrs:
  - "192.168.32.0/24"
  - "127.0.0.0/8"

log:
  level: "info"
  audit_retention_days: 180
  security_log_path: "/var/log/homeshare/security.log"
  security_log_retention_days: 7

speed_limits:
  external_upload_mbps: 250
  external_download_mbps: 250
  burst_mb: 16

quotas:
  default_storage_quota_bytes: 107374182400      # 100 GB
  default_monthly_upload_bytes: 214748364800     # 200 GB
  default_monthly_download_bytes: 322122547200   # 300 GB
  default_max_file_size_bytes: 53687091200       # 50 GB
  default_max_concurrent_uploads: 1

traffic:
  upload_wasted_allowance: 0.5
  download_wasted_allowance: 1.0
  grace_factor: 0.5

rate_limits:
  admin_login_failed_per_15min: 5
  admin_totp_failed_per_15min: 5
  invite_failed_per_15min: 5
  upload_create_per_hour_person: 10
  chunk_per_minute_person: 120
  download_per_minute_person: 120
  concurrent_downloads_per_person: 2
  zip_per_hour_person: 5

suspicious_extensions:
  - exe - msi - msp - bat - cmd - com - scr - vbs - vbe
  - js - jse - ws - wsf - wsh - ps1 - psm1 - sh - bash
  - dll - ocx - jar - apk - hta - cpl

zip:
  max_files: 100
  max_total_bytes: 53687091200  # 50 GB
  store_mode: true

sessions:
  user_idle_days: 30
  user_absolute_days: 90
  admin_idle_hours: 12
  admin_absolute_days: 7

invite:
  default_expires_hours: 24
  default_max_activations: 1

disk:
  min_free_space_gb: 40
  critical_free_space_gb: 20
  min_free_inodes: 100000

upload:
  chunk_size_bytes: 33554432  # 32 MB
  min_reservation_ttl_hours: 1
  max_reservation_ttl_hours: 72
  ttl_factor: 2.0
  default_estimated_speed_bps: 2097152

security:
  quarantine_suspicious: true
  session_secret_env: "HOMESHARE_SESSION_SECRET"
  ip_hash_salt_env: "HOMESHARE_IP_HASH_SALT"

cleanup:
  run_interval_seconds: 60
  backup_daily: true
  backup_keep_days: 14
```

## Systemd Unit

```ini
[Unit]
Description=Homeshare File Sharing Service
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=homeshare
Group=homeshare
ExecStart=/usr/local/bin/homeshare -config /etc/homeshare/config.yaml
Restart=on-failure
RestartSec=5
WorkingDirectory=/var/lib/homeshare

# Security hardening
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=read-only
PrivateTmp=true
ReadWritePaths=/srv/media/fileshare /home/fileshare-backup /var/log/homeshare

# Environment
Environment="GODEBUG=asyncpreemptoff=1"
Environment="HOMESHARE_SESSION_SECRET="
Environment="HOMESHARE_IP_HASH_SALT="

[Install]
WantedBy=multi-user.target
```

## Caddyfile

```caddy
# Внешний доступ через DuckDNS
files.example.duckdns.org {
    reverse_proxy 127.0.0.1:8090 {
        header_up X-Real-IP {http.request.remote.host}
        header_up X-Forwarded-Proto {scheme}
        header_up X-Forwarded-For {http.request.remote.host}
    }
}

# Локальный доступ по IP
https://192.168.32.10 {
    tls internal
    
    reverse_proxy 127.0.0.1:8090 {
        header_up X-Real-IP {http.request.remote.host}
        header_up X-Forwarded-Proto {scheme}
    }
}
```

## Fail2ban Filter

`/etc/fail2ban/filter.d/homeshare.conf`:

```ini
[Definition]
failregex = ^.*(?:admin_login_failed|admin_totp_failed|invite_failed|rate_limited).*<HOST>.*$
ignoreregex =
```

`/etc/fail2ban/jail.d/homeshare.local`:

```ini
[homeshare]
enabled = true
filter = homeshare
logpath = /var/log/homeshare/security.log
maxretry = 5
bantime = 3600
findtime = 900
```

## Инструкция по установке

### 1. Подготовка системы

```bash
# Создать пользователя
sudo useradd --system --shell /bin/false --home-dir /var/lib/homeshare homeshare

# Создать директории
sudo mkdir -p /srv/media/fileshare/{data,tmp,db}
sudo mkdir -p /home/fileshare-backup
sudo mkdir -p /var/log/homeshare
sudo mkdir -p /etc/homeshare
sudo mkdir -p /var/lib/homeshare

# Установить права
sudo chown -R homeshare:homeshare /srv/media/fileshare
sudo chown -R homeshare:homeshare /home/fileshare-backup
sudo chown -R homeshare:homeshare /var/log/homeshare
sudo chown -R homeshare:homeshare /var/lib/homeshare
sudo chmod 750 /srv/media/fileshare/{data,tmp,db}
sudo chmod 750 /home/fileshare-backup
sudo chmod 750 /var/log/homeshare
```

### 2. Установить Caddy

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

### 3. Настроить Caddy

```bash
sudo cp Caddyfile /etc/caddy/Caddyfile
sudo systemctl restart caddy
```

### 4. Установить homeshare

```bash
# Скомпилировать
cd /workspace
go build -o homeshare ./cmd/homeshare

# Установить
sudo cp homeshare /usr/local/bin/
sudo chmod 755 /usr/local/bin/homeshare
sudo cp config.yaml /etc/homeshare/config.yaml
sudo chmod 600 /etc/homeshare/config.yaml
sudo chown homeshare:homeshare /etc/homeshare/config.yaml
```

### 5. Сгенерировать секреты

```bash
# Сгенерировать секреты
SESSION_SECRET=$(openssl rand -hex 32)
IP_HASH_SALT=$(openssl rand -hex 32)

# Добавить в systemd environment или .env файл
echo "HOMESHARE_SESSION_SECRET=$SESSION_SECRET" | sudo tee -a /etc/homeshare/env
echo "HOMESHARE_IP_HASH_SALT=$IP_HASH_SALT" | sudo tee -a /etc/homeshare/env
sudo chmod 600 /etc/homeshare/env
sudo chown homeshare:homeshare /etc/homeshare/env
```

### 6. Установить systemd unit

```bash
sudo cp homeshare.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable homeshare
```

### 7. Создать первого админа

```bash
sudo -u homeshare homeshare admin create --username admin
# Следовать инструкциям для установки пароля и TOTP
```

### 8. Запустить сервис

```bash
sudo systemctl start homeshare
sudo systemctl status homeshare
```

### 9. Настроить UFW

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
# НЕ открывать 8090 наружу!
```

### 10. Настроить DuckDNS

1. Зарегистрироваться на duckdns.org
2. Создать subdomain: files.example.duckdns.org
3. Настроить обновление IP (скрипт или Docker контейнер)
4. Пробросить порты 80/443 на сервер

## CLI Команды

```bash
# Создать админа
homeshare admin create --username <name>

# Удалить админа
homeshare admin delete --username <name>

# Сбросить TOTP
homeshare admin reset-totp --username <name>

# Разблокировать админа
homeshare admin unlock --username <name>

# Показать версию
homeshare --version

# Запустить с кастомным конфигом
homeshare -config /path/to/config.yaml
```

## Checklist проверки

- [ ] Сервис слушает только 127.0.0.1:8090
- [ ] Caddy терминирует TLS
- [ ] X-Real-IP устанавливается корректно
- [ ] Локальные клиенты (192.168.32.0/24) не лимитируются
- [ ] Внешние клиенты лимитируются по скорости
- [ ] Invite коды работают
- [ ] TOTP для админов работает
- [ ] Загрузка файлов работает (chunked)
- [ ] Скачивание файлов работает (range requests)
- [ ] Quarantine подозрительных файлов работает
- [ ] Traffic counters считаются правильно
- [ ] Grace rule применяется
- [ ] Rate limits работают
- [ ] Audit log пишется
- [ ] Security log пишется с raw IP
- [ ] Fail2ban банит по security log
- [ ] Backup БД создаётся ежедневно
- [ ] Очистка expired файлов работает
- [ ] Disk reserve проверяется

## Тестирование curl

```bash
# Проверить что сервис отвечает
curl http://127.0.0.1:8090/

# Активировать invite (пример)
curl -X POST http://127.0.0.1:8090/api/invite/activate \
  -H "Content-Type: application/json" \
  -d '{"code": "XXXX-XXXX-XXXX-XXXX"}'

# Создать upload
curl -X POST http://127.0.0.1:8090/api/uploads \
  -H "Content-Type: application/json" \
  -H "Cookie: session=..." \
  -d '{"filename": "test.txt", "size": 1024, "content_type": "text/plain", "expiry_days": 14}'

# Загрузить chunk
curl -X PATCH "http://127.0.0.1:8090/api/uploads/1?offset=0" \
  -H "Authorization: Bearer <upload_secret>" \
  --data-binary @chunk1.bin

# Завершить upload
curl -X POST http://127.0.0.1:8090/api/uploads/1/complete \
  -H "Authorization: Bearer <upload_secret>"

# Скачать файл
curl -O -J http://127.0.0.1:8090/api/files/<file_id> \
  -H "Cookie: session=..."
```

## Безопасность

- Все пароли хешируются Argon2id
- TOTP обязательна для админов
- Session tokens хранятся только как hash
- Invite codes хранятся только как hash
- IP адреса хешируются в audit log
- Security log содержит raw IP только для fail2ban
- Файлы хранятся под случайными именами
- Path traversal защищён
- CSP headers установлены
- X-Content-Type-Options: nosniff
- HttpOnly cookies
- Secure cookies при HTTPS
