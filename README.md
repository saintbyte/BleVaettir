# BleVaettir

BLE сенсорный демон для Linux. Собирает данные с BLE устройств, обрабатывает и сохраняет/отправляет.

## Содержание

- [Установка](#установка)
- [Конфигурация](#конфигурация)
- [Запуск](#запуск)
- [BLE объекты](#ble-объекты)
- [Парсеры](#парсеры)
- [Обработчики (handlers)](#обработчики-handlers)
- [Неизвестные устройства](#неизвестные-устройства)
- [Сборка](#сборка)
- [Systemd](#systemd)
- [Примеры](#примеры)
- [Дополнение](#Дополнение)
## Установка

### Из бинарников

```bash
# Скачать с релизов
# amd64
sudo cp bin/linux/amd64/blevaettir /usr/local/bin/

# arm64
sudo cp bin/linux/arm64/blevaettir /usr/local/bin/

# Создать директории
sudo mkdir -p /etc/blevaettir /var/lib/blevaettir /var/log

# Скопировать конфиг
sudo cp blevaettir.yaml.example /etc/blevaettir/blevaettir.yaml

# Установить права
sudo chown root:root /usr/local/bin/blevaettir
sudo chmod +x /usr/local/bin/blevaettir
```

### Из исходников

```bash
git clone https://github.com/saintbyte/BleVaettir.git
cd BleVaettir
make build
```

## Конфигурация

Конфигурация хранится в YAML файле. Полный пример: [`blevaettir.yaml.example`](./blevaettir.yaml.example)

### Структура конфига

```yaml
ble:
  hci: 0                    # HCI интерфейс Bluetooth (обычно 0)

storage:
  path: "/var/lib/blevaettir/data.db"  # Путь к SQLite базе

intervals:
  scan_duration_sec: 5      # Длительность одного сканирования (секунды)
  scan_interval_sec: 30     # Интервал между сканированиями (секунды)

log:
  level: "info"             # Уровень логирования: debug, info, warn, error

handlers:                   # Глобальные обработчики (применяются к объектам без собственных)
  - type: "db"
    db:
      enabled: true
  - type: "http"
    http:
      enabled: false
      endpoint: "http://localhost:8080/api/sensors"
      api_key: ""
      ca_cert: ""           # Путь к CA сертификату (опционально)
      client_cert: ""        # Путь к клиентскому сертификату (опционально)
      client_key: ""         # Путь к ключу клиента (опционально)
      skip_verify: false     # Пропустить проверку TLS (не рекомендуется)
  - type: "log"
  - type: "narodmon"
    narodmon:
      enabled: false
      endpoint: "http://narodmon.ru/json"
      owner: "Your Name"
      lat: "55.7558"
      lon: "37.6173"
      alt: "150"
  - type: "datacake"
    datacake:
      enabled: false
      endpoint: "https://api.datacake.co"
      skip_verify: false

unknown_objects:
  enabled: true             # Обрабатывать неизвестные устройства
  handlers:
    - type: "log"

ble_objects:                # Список отслеживаемых BLE устройств
  - name: "Kitchen"
    mac: "A4:C1:38:12:34:56"
    parsers:
      - type: "xiaomi_lywsd03mmc"
    handlers:               # Собственные обработчики (опционально)
      - type: "db"
```

### Параметры секции `ble`

| Параметр | Тип | Описание | По умолчанию |
|----------|-----|----------|--------------|
| `hci` | int | HCI интерфейс Bluetooth | `0` |

### Параметры секции `storage`

| Параметр | Тип | Описание | По умолчанию |
|----------|-----|----------|--------------|
| `path` | string | Путь к SQLite базе данных | `""` |

### Параметры секции `intervals`

| Параметр | Тип | Описание | По умолчанию |
|----------|-----|----------|--------------|
| `scan_duration_sec` | int | Длительность одного сканирования (сек) | `5` |
| `scan_interval_sec` | int | Интервал между сканированиями (сек) | `30` |

### Параметры секции `log`

| Параметр | Тип | Описание | По умолчанию |
|----------|-----|----------|--------------|
| `level` | string | Уровень логирования: debug, info, warn, error | `info` |

## Запуск

### Режимы работы

```bash
# Интерактивный режим (вывод в консоль)
blevaettir -config /etc/blevaettir/blevaettir.yaml

# Демон (отключается от консоли)
blevaettir -d -config /etc/blevaettir/blevaettir.yaml

# С логированием в файл
blevaettir -d -config /etc/blevaettir/blevaettir.yaml -log-file /var/log/blevaettir.log
```

### Флаги командной строки

| Флаг | Описание | По умолчанию |
|------|----------|--------------|
| `-config` | Путь к конфигу | `blevaettir.yaml` |
| `-d` | Запуск как демон | `false` |
| `-log-file` | Файл логов (для демона) | stdout |

## BLE объекты

`ble_objects` - список устройств для мониторинга.

### Структура

```yaml
ble_objects:
  - name: "Имя устройства"     # Человеческое имя
    mac: "AA:BB:CC:DD:EE:FF"  # MAC адрес (обязательно)
    parsers:                    # Парсеры данных (минимум 1)
      - type: "xiaomi_lywsd03mmc"
    handlers:                   # Собственные обработчики (опционально)
      - type: "db"
        db:
          enabled: true
```

### Параметры объекта

| Параметр | Тип | Обязательный | Описание |
|----------|-----|--------------|----------|
| `name` | string | Да | Человеческое имя устройства |
| `mac` | string | Да | MAC адрес устройства |
| `parsers` | []object | Да | Список парсеров |
| `handlers` | []object | Нет | Собственные обработчики (если указаны, глобальные игнорируются) |

## Парсеры

Парсер определяет формат данных из advertisement пакета.

### Типы парсеров

| Тип | Описание | Данные |
|-----|----------|--------|
| `xiaomi_lywsd03mmc` | Xiaomi LYWSD03MMC | temperature, humidity |
| `atc_thermometer` | ATC/BLE термометры | temperature, humidity |
| `jaalee` | Jaalee датчики | temperature, humidity, battery, rssi |
| `raw` | Сырые данные | raw (первый байт) |

### Xiaomi LYWSD03MMC

```yaml
parsers:
  - type: "xiaomi_lywsd03mmc"
```
Выводит:
- `temperature` - температура (°C)
- `humidity` - влажность (%)

### ATC Thermometer

```yaml
parsers:
  - type: "atc_thermometer"
```
Выводит:
- `temperature` - температура (°C)
- `humidity` - влажность (%)

### Jaalee Thermometer

```yaml
parsers:
  - type: "jaalee"
```
Выводит:
- `temperature` - температура (°C)
- `humidity` - влажность (%)
- `battery` - заряд батареи (%)
- `rssi` - уровень сигнала (dBm)

### Raw

```yaml
parsers:
  - type: "raw"
```
Выводит значение первого байта manufacturer data.

### Множественные парсеры

К одному устройству можно применить несколько парсеров:

```yaml
ble_objects:
  - name: "Test Device"
    mac: "AA:BB:CC:DD:EE:FF"
    parsers:
      - type: "xiaomi_lywsd03mmc"
      - type: "raw"
```

## Обработчики (handlers)

Обработчики определяют что делать с полученными данными.

### Типы обработчиков

| Тип | Описание |
|-----|---------|
| `db` | Сохранение в SQLite |
| `http` | HTTP POST запрос |
| `log` | Логирование через slog |
| `narodmon` | Отправка на Narodmon.ru |
| `datacake` | Отправка в DataCake |

### Глобальные и объектные обработчики

**Глобальные обработчики** (`handlers`) применяются ко всем устройствам без собственных обработчиков:

```yaml
handlers:
  - type: "db"
    db:
      enabled: true
  - type: "log"
```

**Объектные обработчики** переопределяют глобальные для конкретного устройства:

```yaml
ble_objects:
  - name: "Kitchen"
    mac: "A4:C1:38:12:34:56"
    parsers:
      - type: "xiaomi_lywsd03mmc"
    handlers:
      - type: "db"
        db:
          enabled: true
      - type: "log"
```

### Обработчик DB

Сохранение показаний в SQLite базу.

```yaml
handlers:
  - type: "db"
    db:
      enabled: true
```

**Таблица `readings`:**

| Поле | Тип | Описание |
|------|-----|----------|
| `id` | INTEGER | PRIMARY KEY |
| `sensor_mac` | TEXT | MAC адрес |
| `sensor_name` | TEXT | Имя из конфига |
| `type` | TEXT | Тип данных (temperature, humidity, raw) |
| `value` | REAL | Значение |
| `unit` | TEXT | Единица измерения |
| `timestamp` | TIMESTAMP | Время измерения |

### Обработчик HTTP

Отправка данных HTTP POST запросом в JSON формате.

```yaml
handlers:
  - type: "http"
    http:
      enabled: true
      endpoint: "http://localhost:8080/api/sensors"
      api_key: "secret_key"           # Bearer token (опционально)
      ca_cert: "/path/to/ca.pem"      # CA сертификат (опционально)
      client_cert: "/path/to/cert.pem" # Клиентский сертификат (опционально)
      client_key: "/path/to/key.pem"   # Ключ клиента (опционально)
      skip_verify: false              # Пропустить проверку TLS
```

**Формат POST запроса:**

```json
[
  {
    "sensor_mac": "A4:C1:38:12:34:56",
    "sensor_name": "Kitchen",
    "type": "temperature",
    "value": 22.5,
    "unit": "°C",
    "timestamp": "2026-03-23T12:00:00Z"
  }
]
```

**Заголовки:**
- `Content-Type: application/json`
- `Authorization: Bearer <api_key>` (если указан)

### Обработчик Log

Простое логирование через slog.

```yaml
handlers:
  - type: "log"
```

Вывод (level=debug):
```
level=INFO msg="sensor reading" mac=A4:C1:38:12:34:56 name=Kitchen type=temperature value=22.5 unit="°C"
```

### Обработчик Narodmon

Отправка данных на сервис [Narodmon.ru](https://narodmon.ru).

```yaml
handlers:
  - type: "narodmon"
    narodmon:
      enabled: true
      endpoint: "http://narodmon.ru/json"
      owner: "Your Name"
      lat: "55.7558"
      lon: "37.6173"
      alt: "150"
```

**Параметры:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `enabled` | bool | Включить обработчик |
| `endpoint` | string | URL API Narodmon |
| `owner` | string | Имя владельца |
| `lat` | string | Широта |
| `lon` | string | Долгота |
| `alt` | string | Высота над уровнем моря |

### Обработчик DataCake

Отправка данных в сервис [DataCake](https://datacake.co).

```yaml
handlers:
  - type: "datacake"
    datacake:
      enabled: true
      endpoint: "https://api.datacake.co"
      skip_verify: false
```

**Параметры:**

| Параметр | Тип | Описание |
|----------|-----|----------|
| `enabled` | bool | Включить обработчик |
| `endpoint` | string | URL API DataCake |
| `skip_verify` | bool | Пропустить проверку TLS |

## Неизвестные устройства

Устройства, обнаруженные при сканировании, но не присутствующие в `ble_objects`.

```yaml
unknown_objects:
  enabled: true
  handlers:
    - type: "log"
    - type: "narodmon"
      narodmon:
        enabled: false
        endpoint: "http://narodmon.com/post"
```

При обнаружении отправляют:
- `sensor_mac` - MAC адрес
- `sensor_name` - "unknown"
- `type` - "raw"
- `value` - количество байт manufacturer data
- `unit` - "bytes"

## Сборка

### Makefile

```bash
# Собрать все платформы
make build

# Очистить
make clean

# Пересобрать
make clean && make build
```

### Платформы

- `linux/amd64` - 64-бит Intel/AMD
- `linux/arm64` - 64-бит ARM
- `linux/386` - 32-бит Intel

### Бинарники

```
bin/
├── amd64/blevaettir
├── arm64/blevaettir
└── 386/blevaettir
```

## Systemd

### Установка

```bash
sudo cp bin/linux/amd64/blevaettir /usr/local/bin/
sudo cp blevaettir.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable blevaettir
sudo systemctl start blevaettir
```

### Управление

```bash
# Статус
sudo systemctl status blevaettir

# Логи
sudo journalctl -u blevaettir -f

# Перезапуск
sudo systemctl restart blevaettir

# Остановка
sudo systemctl stop blevaettir
```

### Сервис файл

См. [`blevaettir.service`](./blevaettir.service)

## Примеры

### Минимальный конфиг

```yaml
ble:
  hci: 0

storage:
  path: "/var/lib/blevaettir/data.db"

intervals:
  scan_duration_sec: 5
  scan_interval_sec: 30

log:
  level: "info"

handlers:
  - type: "db"
    db:
      enabled: true

ble_objects:
  - name: "Thermometer"
    mac: "A4:C1:38:12:34:56"
    parsers:
      - type: "xiaomi_lywsd03mmc"
```

### Несколько устройств, разные обработчики

```yaml
ble:
  hci: 0

storage:
  path: "/var/lib/blevaettir/data.db"

intervals:
  scan_duration_sec: 5
  scan_interval_sec: 30

log:
  level: "info"

handlers:
  - type: "db"
    db:
      enabled: true
  - type: "http"
    http:
      enabled: false
      endpoint: "http://homeassistant:8123/api/sensors"
      api_key: ""

ble_objects:
  - name: "Kitchen"
    mac: "A4:C1:38:11:11:11"
    parsers:
      - type: "xiaomi_lywsd03mmc"
    handlers:
      - type: "db"
      - type: "http"

  - name: "Living Room"
    mac: "A4:C1:38:22:22:22"
    parsers:
      - type: "atc_thermometer"
    handlers:
      - type: "db"

  - name: "Outdoor"
    mac: "A4:C1:38:33:33:33"
    parsers:
      - type: "xiaomi_lywsd03mmc"
    handlers:
      - type: "log"

unknown_objects:
  enabled: true
  handlers:
    - type: "log"
```

### С отладкой

```yaml
log:
  level: "debug"
```

Логи всех показаний:
```
level=INFO msg="sensor reading" mac=A4:C1:38:11:11:11 name=Kitchen type=temperature value=22.5 unit="°C"
level=INFO msg="sensor reading" mac=A4:C1:38:11:11:11 name=Kitchen type=humidity value=65 unit="%"
```

### Отправка на Narodmon

```yaml
ble:
  hci: 0

storage:
  path: "/var/lib/blevaettir/data.db"

intervals:
  scan_duration_sec: 5
  scan_interval_sec: 30

log:
  level: "info"

handlers:
  - type: "narodmon"
    narodmon:
      enabled: true
      endpoint: "http://narodmon.ru/json"
      owner: "My Home"
      lat: "55.7558"
      lon: "37.6173"
      alt: "150"

ble_objects:
  - name: "Balcony"
    mac: "A4:C1:38:AA:BB:CC"
    parsers:
      - type: "xiaomi_lywsd03mmc"
```

### HTTPS с клиентским сертификатом

```yaml
handlers:
  - type: "http"
    http:
      enabled: true
      endpoint: "https://secure-api.example.com/sensors"
      api_key: "my_token"
      ca_cert: "/etc/blevaettir/ca.pem"
      client_cert: "/etc/blevaettir/client.pem"
      client_key: "/etc/blevaettir/client.key"
```

## Требования

- Linux с BlueZ
- Bluetooth адаптер с поддержкой BLE
- Go 1.25+

## Устранение неполадок

### BLE не работает

```bash
# Проверить HCI интерфейс
hciconfig -a

# Сбросить адаптер
sudo hciconfig hci0 down
sudo hciconfig hci0 up

# Запустить от root
sudo ./blevaettir -config config.yaml
```

### Нет данных от сенсора

1. Убедитесь что MAC адрес правильный
2. Проверьте поддерживаемый тип парсера
3. Включите debug логирование
4. Проверьте что сенсор вещает advertise пакеты

### Ошибка "failed to open HCI"

```bash
# Проверить права доступа
ls -l /dev/hci*
groups $USER
sudo usermod -a -G bluetooth $USER
# Перелогиниться
```


# Дополнение
1. [BleVaettirDbPlot](https://github.com/saintbyte/BleVaettirDbPlot) - тулза чтобы строить графики по данным БД
