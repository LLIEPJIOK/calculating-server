## Calculation server

Читайте это на других языках: [English](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/README.md).

Calculation server - это сервер, который вычисляет выражения в течение заданного времени операции.

В проекте также используется база данных PostgresSQL, [Bootstrap](https://getbootstrap.com) и [HTMX](https://htmx.org). Bootstrap используется для создания быстрого и визуально привлекательного интерфейса, в то время как HTMX используется для отправки запросов и их удобной обработки.

## Начало работы
Для запуска сервера вам понадобится только [Docker](https://www.docker.com/products/docker-desktop/), запущенный на вашем компьютере.

Чтобы начать работу с сервером, выполните следующие шаги:
1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/LLIEPJIOK/calculating-server.git
   ```
2. Перейдите в файл проекта:
   ```bash
   cd calculating-server
   ```
3. Запустите докер контейнер с проектом:
   ```bash
   docker-compose up
   ```
4. Откройте [`localhost:8080`](http://localhost:8080) в вашем браузере.

Если вы запустили приложение в первый раз, вам необходимо будет зарегестрироваться. Иначе просто войдите. После этого сервер выдаст вам доступ для применения операций над выражениями и запомнит вас на 1 день. Теперь можете выполнить желаемые операции и увидеть результаты.

*Примечание: Когда запускается контейнер отрабатывают все тесты, поэтому если какой-то тест не пройдёт, то не запуститься и вся программа.*

## Структура кода
1. `main.go` - файл для инициализации проекта.
2. `internal/controllers` - папка с файлами для сервера, где обрабатываются запросы.
3. `internal/database` - папка с файлами для взаимодействия с базой данных PostgreSQL.
4. `internal/expression` - папка с файлами для обработки выражений.
5. `internal/user` - папка с файлами для обработки пользователей.
6. `internal/workers` - папка с файлами для обработки рабочих потоков.
7. `static/` - папка для визуального интерфейса. Она содержит стили CSS, JS-скрипты, иконки и HTML-шаблоны с Bootstrap и HTMX.

## Принцип работы
Перед каждым запросом пользователя (например, открытие страницы, отправка выражения и т. д.) сервер проверяет авторизацию. Если пользователь авторизован, то сервер предоставляет ему доступ для выполнения операций с выражениями, но не для входа в систему или регистрации. В противном случае, сервер позволяет только войти в систему или зарегистрироваться. Аутентификация была реализована с использованием JWT.

После этого сервер обрабатывает запрос, и существует несколько возможных сценариев:

1. **Запрос на вычисление выражения:**
   
   Сервер добавляет выражение в базу данных, затем анализирует его, чтобы убедиться в его правильности.
   - Если выражение допустимо, сервер передает его агентам через канал. В какой-то момент один из агентов забирает его, вычисляет выражение, а затем обновляет результат.
   - Если выражение недопустимо, сервер устанавливает ошибку разбора в статус выражения.
   Затем сервер обновляет последние выражения.

2. **Запрос на обновление времени операции:**

   Сервер обновляет эту информацию в базе данных.

3. **Войти или зарегистрироваться**

   Сервер проверяет правильность данных. Если данные верны, то сервер предоставляет вам доступ для выполнения операций с выражениями и записывает ваши данные в cookie на 1 день, чтобы помнить, что вы зарегистрированы, в противном случае он показывает ошибки.

4. **Другие запросы:**

   Сервер извлекает информацию из базы данных и отображает ее.

![Схема работы](https://github.com/LLIEPJIOK/calculating-server/blob/master/images/WorkingScheme.jpg)

*Примечание: В настоящее время нет автоматического обновления данных на странице. Чтобы увидеть изменения, вы должны перезагрузить страницу.*