# BleVaettir

BLE сенсорный демон для Linux. Собирает данные с BLE устройств, обрабатывает и сохраняет/отправляет.

## Содержание

- [Установка](#установка)
- [Конфигурация](#конфигурация)
- [Запуск](#запуск)
- [BLE объекты](#ble-объекты)
- [Обработчики (handlers)](#обработчики-handlers)
- [Неизвестные устройства](#неизвестные-устройства)
- [Сборка](#сборка)
- [Systemd](#systemd)
- [Примеры](#примеры)

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

Конфигурация хранится в YAML файле.

### Структура конфига

```yaml
ble:
  hci: 0                    # HCI интерфейс Bluetooth (обычно 0)

storage:
  path: "/var/lib/blevaettir/data.db"  # Путь к SQLite базе

intervals:
  scan_interval_sec: 30     # Интервал между сканированиями (секунды)

log:
  level: "info"             # Уровень логирования: debug, info, warn, error

handlers:                   # Глобальные обработчики
  - type: "db"             # Обязательно: сохранение в БД
    db:
      enabled: true
  - type: "http"           # Опционально: отправка по HTTP
    http:
      enabled: false
      endpoint: "http://localhost:8080/api/sensors"
      api_key: ""
  - type: "log"           # Опционально: логирование

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

### Полный пример конфига

См. [`blevaettir.yaml.example`](./blevaettir.yaml.example)

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
```

### Парсеры (parsers)

Парсер определяет формат данных из advertisement пакета.

| Тип | Описание | Данные |
|-----|----------|--------|
| `xiaomi_lywsd03mmc` | Xiaomi LYWSD03MMC | temperature, humidity |
| `atc_thermometer` | ATC/BLE термометры | temperature, humidity |
| `raw` | Сырые данные | raw (первый байт) |

#### Xiaomi LYWSD03MMC

```yaml
parsers:
  - type: "xiaomi_lywsd03mmc"
```
Выводит:
- `temperature` - температура (°C)
- `humidity` - влажность (%)

#### ATC Thermometer

```yaml
parsers:
  - type: "atc_thermometer"
```
Выводит:
- `temperature` - температура (°C)
- `humidity` - влажность (%)

#### Raw

```yaml
parsers:
  - type: "raw"
```
Выводит значение первого байта manufacturer data.

### Множественные парсеры

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

| Тип | Описание | Конфиг |
|-----|----------|---------|
| `db` | Сохранение в SQLite | `db.enabled: true/false` |
| `http` | HTTP POST запрос | `http.endpoint`, `http.api_key` |
| `log` | slog.Info логирование | нет параметров |

### Глобальные обработчики

Применяются ко всем `ble_objects` без собственных обработчиков.

```yaml
handlers:
  - type: "db"
    db:
      enabled: true
  - type: "log"
```

### Собственные обработчики объекта

```yaml
ble_objects:
  - name: "Kitchen"
    mac: "A4:C1:38:12:34:56"
    parsers:
      - type: "xiaomi_lywsd03mmc"
    handlers:
      - type: "db"           # Только в БД
      - type: "log"          # И логировать
```

### Обработчик DB

Сохранение в SQLite базу.

```yaml
handlers:
  - type: "db"
    db:
      enabled: true
```

Таблица `readings`:
- `id` - INTEGER PRIMARY KEY
- `sensor_mac` - MAC адрес
- `sensor_name` - Имя из конфига
- `type` - тип данных (temperature, humidity, raw)
- `value` - значение
- `unit` - единица измерения
- `timestamp` - время

### Обработчик HTTP

Отправка данных POST запросом.

```yaml
handlers:
  - type: "http"
    http:
      enabled: true
      endpoint: "http://localhost:8080/api/sensors"
      api_key: "secret_key"
```

Формат POST запроса (JSON массив):
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

Заголовки:
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

## Неизвестные устройства

Устройства, обнаруженные при сканировании, но не присутствующие в `ble_objects`.

```yaml
unknown_objects:
  enabled: true
  handlers:
    - type: "log"
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

## Требования

- Linux с BlueZ
- Bluetooth адаптер с поддержкой BLE
- Go 1.22+

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
