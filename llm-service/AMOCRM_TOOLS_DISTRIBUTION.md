# Распределение инструментов AmoCRM по агентам

## Общая концепция

Инструменты AmoCRM распределены между агентами согласно их ролям и задачам:
- **Main Agent** - быстрые операции для пользователя без необходимости открывать CRM
- **Business Analyst** - полный read-only доступ для аналитики + создание заметок с инсайтами
- **Marketing Agent** - управление кампаниями через создание и редактирование лидов/контактов
- **Legal Agent** - работает только с контрактами, CRM не использует

Все инструменты используют префикс `ammo-crm-`.

---

## Main Agent (основной ассистент)

**Назначение**: Быстрые повседневные операции без перехода в CRM

### Инструменты чтения:
- `ammo-crm-entity_get` - список сущностей (leads, contacts, companies, tasks)
- `ammo-crm-entity_id_get` - получение конкретной сущности по ID
- `ammo-crm-users_get` - список пользователей
- `ammo-crm-leads_pipelines_get` - список воронок продаж

### Инструменты создания:
- `ammo-crm-entity_notes_post` - создание заметок к сущностям
- `ammo-crm-contacts_post` - создание контактов
- `ammo-crm-tasks_post` - создание задач

### Утилиты:
- `ammo-crm-timestamp_shift_get` - работа с временными метками

**Итого**: 8 инструментов

---

## Business Analyst (бизнес-аналитик)

**Назначение**: Глубокая аналитика данных CRM без возможности редактирования

### Инструменты чтения (полный доступ):
- `ammo-crm-entity_get` - список сущностей с фильтрацией
- `ammo-crm-entity_id_get` - детальная информация по сущности
- `ammo-crm-entity_notes_get` - чтение заметок
- `ammo-crm-entity_custom_fields_get` - список кастомных полей
- `ammo-crm-entity_custom_fields_id_get` - информация о конкретном кастомном поле
- `ammo-crm-entity_links_get` - связи между сущностями
- `ammo-crm-leads_pipelines_get` - список воронок
- `ammo-crm-leads_pipelines_id_get` - детали конкретной воронки
- `ammo-crm-leads_pipelines_id_statuses_get` - статусы воронки
- `ammo-crm-leads_pipelines_id_statuses_status_id_get` - детали конкретного статуса
- `ammo-crm-users_get` - список пользователей/менеджеров
- `ammo-crm-tasks_types_get` - типы задач

### Инструменты создания (ограниченно):
- `ammo-crm-entity_notes_post` - создание заметок для фиксации инсайтов и рекомендаций

### Утилиты:
- `ammo-crm-timestamp_shift_get` - работа с датами для аналитики периодов

**Итого**: 14 инструментов (13 read + 1 create для заметок)

**Важно**: НЕТ доступа к редактированию (patch) и удалению данных

---

## Marketing Agent (маркетолог)

**Назначение**: Управление маркетинговыми кампаниями, лидами и контактами

### Инструменты чтения:
- `ammo-crm-entity_get` - поиск и фильтрация сущностей
- `ammo-crm-entity_id_get` - просмотр деталей
- `ammo-crm-entity_links_get` - просмотр связей между сущностями
- `ammo-crm-entity_custom_fields_get` - работа с кастомными полями
- `ammo-crm-leads_pipelines_get` - просмотр воронок
- `ammo-crm-users_get` - список менеджеров

### Инструменты создания:
- `ammo-crm-leads_post` - создание новых сделок
- `ammo-crm-contacts_post` - создание контактов
- `ammo-crm-companies_post` - создание компаний
- `ammo-crm-tasks_post` - постановка задач команде
- `ammo-crm-entity_notes_post` - добавление заметок
- `ammo-crm-entity_links_post` - создание связей между сущностями

### Инструменты редактирования:
- `ammo-crm-leads_patch` - изменение параметров сделок
- `ammo-crm-contacts_patch` - обновление контактной информации
- `ammo-crm-tasks_patch` - изменение задач

### Утилиты:
- `ammo-crm-timestamp_shift_get` - работа с датами для кампаний

**Итого**: 16 инструментов

---

## Legal Agent (юрист)

**Инструменты AmoCRM**: не используются

Агент работает исключительно с инструментами генерации контрактов:
- `search_contract_templates`
- `generate_contract`
- `list_generated_contracts`

---

## Технические детали

### Префикс инструментов
Все инструменты AmoCRM используют префикс `ammo-crm-` (обратите внимание на двойное "m").

### Поддержка паттернов
Система поддерживает wildcard-паттерны (например, `ammo-crm-*`), но для точного контроля доступа используются явные списки инструментов.

### Проверка доступа
Проверка происходит через функцию `MatchesToolPattern` в `domain/agent.go`, которая сопоставляет запрашиваемый инструмент со списком разрешённых для агента.

### Файл конфигурации
Все определения агентов находятся в:
```
llm-service/internal/service/agent/registry.go
```
