# Сервис динамического сегментирования пользователей

# Запуск

~~~zsh
git clone https://github.com/realPointer/segments
cd segments
make compose-up
~~~

Для запуска сервиса с интеграцией с Yandex Disk нужно:
- Получить OAuth-токен [Тык!](https://yandex.ru/dev/disk/poligon/)
- Вписать его в переменную окружения **YANDEX_TOKEN** в `.env`

# Swagger

После запуска приложения доступна Swagger-документация по адресу [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

![swagger](https://github.com/realPointer/segments/assets/50529632/93f20d6b-ccbe-4fd8-a73c-7ba19c1e15e4)


# Запросы

### Создание пользователя

~~~zsh
curl --location --request POST 'localhost:8080/v1/user/{user_id}'
~~~

---

### Получение сегментов пользователя

~~~zsh
curl --location 'localhost:8080/v1/user/{user_id}/segments'
~~~

Пример ответа:
~~~json
[
    "Segment1",
    "Segment2",
    "Segment3"
]
~~~

---

### Создание сегмента

`?auto={percentage}` - опциональный параметр. Лучше туда передать число >0 и <=100

Без него сегмент просто добавится. Дальше можно будет привязать его к какому-нибудь пользователю самостоятельно

~~~zsh
curl --location --request POST 'localhost:8080/v1/segment/{segment_name}?auto={percentage}'
~~~

---

### Удаление сегмента

При удалении сегмента он будет отвязан у всех пользователей. Это запишется в историю каждого пользователя
~~~zsh
curl --location --request DELETE 'localhost:8080/v1/segment/{segment_name}'
~~~

---

### Получение списка сегментов

~~~zsh
curl --location 'localhost:8080/v1/segment/list'
~~~

Пример ответа:
~~~json
[
    "Segment1",
    "Segment2",
    "Segment3"
]
~~~

### Добавление и удаление сегментов пользователю

--- 

`"expire": "1m"` - опциональный параметр. Через это время сегмент будет удалён

А вот такие единицы измерения он может принять: "ns", "µs", "ms", "s", "m", "h"

Также можно лишь добавить или же удалить сегменты
~~~zsh
curl --location 'localhost:8080/v1/user/{user_id}/segments' \
--header 'Content-Type: application/json' \
--data '{
    "add_segments": [
        {
            "name": "{segment_name}"
        },
        {
            "name": "{segment_name}",
            "expire": "1m"
        }
    ],
    "delete_segments": [
        "{segment_name}"
    ]
}'
~~~

---

### Получение операций пользователя в CSV

`?date={year}-{month}` - опциональный параметр. Без него будет выведена полная история пользователя
~~~zsh
curl --location 'localhost:8080/v1/user/{user_id}/operations?date={year}-{month}'
~~~

Пример ответа:
~~~csv
(1, AVITO, add, 2023-08-31 14:24:33.253191 +0000 UTC)
(1, AVITO_300, add, 2023-08-31 14:24:33.253191 +0000 UTC)
(1, AVITO, delete, 2023-08-31 14:24:33.253191 +0000 UTC)
(1, AVITO_300, delete, 2023-08-31 14:24:54.664546 +0000 UTC)
(1, AVITO, add, 2023-08-31 14:26:18.835481 +0000 UTC)
(1, AVITO, delete, 2023-08-31 14:26:36.639305 +0000 UTC)
(1, TEST_AUTO, add, 2023-08-31 15:51:20.77629 +0000 UTC)
~~~

---

### Получение ссылки на скачивание операций пользователя в CSV

`?date={year}-{month}` - опциональный параметр. Без него в файле будет полная история
~~~zsh
curl --location 'localhost:8080/v1/user/{user_id}/operations/report-link?date={year}-{month}'
~~~

Пример ответа:

Вот это ссылочка! 😱😱😱 Пожалуй, скрою её под текстом :d

[Очень длинная ссылка, которую выдаёт API Яндекс Диска](https://downloader.disk.yandex.ru/disk/4a3e713542172d61b7ac0e42debec3aa6960e0faf9f4133acef03da248ebf0a2/64f0f2c4/Ea6pZ581juK3KgOMTe2aoO_05tBn1_J3dNzkU0k11KlvqUvNNDZ4H01HQcCGGr0cThR0FOLzPtfuYClvCWugiQ%3D%3D?uid=1886155152&filename=1.csv&disposition=attachment&hash=&limit=0&content_type=text%2Fplain&owner_uid=1886155152&fsize=416&hid=0226a47b5dae2fe464d9e7924a9b1ad8&media_type=spreadsheet&tknv=v2&etag=694e72f237b472ba2a725729b67b1016)

## Задания

Основное задание (минимум):

- [x] Метод создания сегмента
- [x] Метод удаления сегмента
- [x] Метод добавления и удаления пользователя в сегмент
- [x] Метод получения активных сегментов пользователя

Доп. задание 1:

- [x] Сохранение истории попадания/выбывания пользователя из сегмента с возможностью получения отчета по пользователю за определенный период

Создана отдельная таблица user_segments_log, в которую записываются действия с каждым сегментом пользователя. Имеется возможность получить отчёт за определённый месяц. 

Так как в задании говорится о получении ссылки, то реализован метод получения отчёта с Яндекс Диска. Для этого были использованы стандартные методы API Диска

Доп. задание 2:

- [x] Реализовать возможность задавать TTL (время автоматического удаления пользователя из сегмента)

Данное задание было сделано через пакет планирования go-cron. Каждую минуту происходит выполнение функции, которая ищет устаревшие записи, удаляет и заносит их в историю.

В целом, эту задачу можно было бы реализовать, например, через простую горутину, триггер в базе данных или же cron как отдельная утилита или pgcron - расширение в базе данных.

Доп. задание 3:

- [x] Добавить опцию указания процента пользователей, которые будут попадать в сегмент автоматически

Добавлен опциональный параметр `?auto={percentage}`. Если его передать, то сегмент автоматически будет привязан к указанному проценту пользователей


## Какие-то дополнительные мысли

- Так как в деталях по заданию было указано, что механизм миграции не нужен, то база создаётся при первичной инициализации репозитория. Итоговые таблицы лежат в schemes/scheme.sql
- Была изначально идея генерировать UID для каждого пользователя. Но так как скорее всего в сервисе база пользователей должна поступать извне, то были сделаны обычные целочисленный id
- Достаточно поздно подумал, что в целом все входные параметры можно передавать JSONом, но в пути даже проще
- Использование транзакций для добавления или удаление сегментов у пользователя, чтобы и история точно записалась
- Так как уровень репозитория написан через интерфейсы, то не будет никаких проблем сгенерировать для них моки
- Стоило бы пользоваться гитом не только для финального коммита. Может вообще в следующий раз попробовать GitFlow 🤔
 
## TODO
- Немного валидации. Хоть и присутствует некоторая по типу введённой даты или числового id
- CI/CD
