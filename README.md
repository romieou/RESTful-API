# Тестовое задание 
# Язык: golang

Для запуска программы наберите следующую команду в консоли:

```
docker compose up
```
Пользователям ubuntu следует использовать "-", также для buildkit можно добавить следующее:
```
DOCKER_BUILDKIT=1 docker-compose up
```
Таким образом вы запускаете REST API на``` http:localhost:8080/``` (см. main.go и docker-compose.yml).

1. Путь ```/rest/substr```

Чтобы найти максимальную подстроку, не содержащую повторяющихся символов, нужно ввести
строку в тело запроса в привычном виде. Для удобства можете воспользоваться postman.
Отправьте POST-запрос по endpoint ```/rest/substr/find```

Например, 
```
"abcdabc"
```
Запрос выше вернет ответ "abcd", так как эта подстрока самая длинная в исходной строке.

| ⚠️WARNING: Заметьте, что программа лишь принимает строку, состоящую исключительно из латинских букв! |
| --- |

При неверном вводе, например, ```"ушаруц"``` или ```"asas182712"```, программа вернет ошибку 400 и соответствующее сообщение ```invalid input```

Реализовано с помощью хендлеров SubstringHandler и GetSubstring. Тесты находятся в файле ```find_substr_test.go```, где можно найти больше примеров применения.

2. Путь ```/rest/email```

Приемлемый формат ввода представлен ниже:
```
«Email:__email@email.com»
```
Программа принимает лишь одну строку, обозначенную двойными кавычками с обеих сторон. Строка может содержать несколько валидных представлений искомого формата, где вместо ```___``` может содержаться несколько нижних пробелов или переносов строки. Перенос строки следует обозначать как "\n".

Например,
```
"Email:__email@gmail.com\nEmail:__\n__\nram.osp98@gmail.com\n__dog$@krispie.hrEmail:__dog@krispie.hr Email:__________________ram.osp98@krispie.hr\n"
```
, где email@gmail.com, ram.osp98@gmail.com, dog@krispie.hr, ram.osp98@krispie.hr являются корректными и возвращаются в виде ответа.

Если не найдено ни единой валидной электронной почты соглано формату, возвращается ошибка 404. В остальных случаях при неправильном вводе выводится ошибка 400.

Реализовано хендлерами EmailHandler и GetEmail.


*Поиск последовательности цифр, являющейся корректным ИИН

Функционал схож с предыдущим пунктом. Алгоритм не применим ко всем родившимся после 2000. Принимается единая строка по endpoint ```/rest/iin/check``` в виде:
```
«IIN:__123456789012»
```

Реализовано с помощью хендлера GetIIN.

Тесты для обоих функционалов прописаны в файле ```email_test.go```.

3. Путь ```/rest/counter```

Простая реализация счетчика, осуществленная хендлерами Add, AddCounter, SubCounter и GetCounter. Тесты приведены в файле counter_test.go. Тесты написаны без поднятия redis благодаря удобству interface в Golang. Счетчик автоматически иницилизируется программой.

* Счетчик можно увеличить, отправив POST-запрос по endpoint ```/rest/counter/add/$i```, где i - целое число, которое прибавляется к текущему значению счетчика, полученного из redis. 

    Можно отправлять и отрицательные числа, и начинать цифру с одним или несколькими нулями (что не очень логично :P)

* Чтобы убавить счетчик на определенное число, отправляйте его как часть пути так же через POST-запрос по ```/rest/counter/sub/$i```.

    Здесь i - число целое, также может принимать отрицательную форму (:P), главное, чтобы отправляемое значение не превышало сам счетчик. В таком случае программа возвращает ошибку 400.

* Наконец, для получения текущего значения счетчика нужно отправить GET-запрос по ```/rest/counter/val```.

4. Путь ```/rest/user```

Реализация CRUD (Create-Read-Update-Delete)-операций над пользователем. У пользователя есть свой ID (генерируемый БД Mysql), имя, фамилия.


* Для создания нового пользователя, нужно отправить POST-запрос по ```/rest/user``` с телом в виде JSON:
```
{
    "first_name": "Latinonly",
    "second_name": "Latinonly"
}
```

Добавив пользователя в базу, сервер возвращает присвоенный ему / ей базой ID:
```
Success! Created new user under ID 5
```

* GET-запрос по ```/rest/user/:id``` поможет извлечь информацию о пользователе с заданным ID (вместо :id). Ответ приходит в виде JSON, но если пользователь не найден, генерируется ошибка 404. При некорректном вводе, например ID, не являющемся цифрой, запрашивающий получает ошибку 400.

Ответ при корректном запросе выглядит следующим образом:
```
{
    "id": 5,
    "first_name": "James",
    "last_name": "McAvoy"
}
```

* Чтобы ввести изменения в данные существующего пользователя, отправляйте PUT-запрос по ```/rest/user/:id```

Так же, как и при создании пользователя, желаемые изменения необходимо приводить в JSON:
```
{
    "first_name": "Latinonly",
    "second_name": "Latinonly"
}
```

При этом одно из полей можно опустить, но не оба:

```
{
    "first_name": "Latinonly"
}
```

или

```
{
    "second_name": "Latinonly"
}
```

Указанные выше варианты оба допустимы. Изменения можно увидеть, лишь пройдя по пункту 4.2.

* Для удаления пользователя из базы необходимо отправить DELETE-запрос по ```/rest/user/:id```, указав лишь ID пользователя. Если таковой существует, вы увидите ответ:
```
Success! Deleted user under ID 5
```

5. Путь ```/rest/hash```

Реализация подсчета следующей хэш-функции:
1) Взять CRC64 хэш от входной строки
2) Взять текущий timestamp с точностью до наносекунд
3) Сделать логическое «И» текущего timestamp и текущего хэша
4) Повторить шаги 2-3 в течение минуты с интервалом в 5 секунд
5) Посчитать число единиц в двоичной записи полученного числа. Количество единиц
и будет являться «хэшом»

* Клиенту необходимо отправить заявку на вычисление hash по методу POST на ```/rest/hash/calc``` с телом в виде простой строки, содержащую CRC64:
```
"11110000111"
```
В ответ пользователь получит номер заявки в следующем образе:
```
We have received your request and assigned the ID 0f0b72d8-e4e1-4746-a36c-126e0f899efd
```
* По окончании вычислений пользователь може увидеть сгенерированный hash методом GET по ```/rest/has/result/$id```, где id - уникальный идентификатор, полученный при отправке запроса:
```
Your hash is 0
```

Пока запрос обрабатывается, пользователя извещают о том, что заявка на рассмотрении:  

```
Your hash is PENDING
```

Все сгенерированные значения хранятся в том же redis, что и использован в задании #3 для хранения счетчика.

    a) Метод GetTimestamp() извлекает текущий timestamp, с учетом того, что в один момент времени ее может вызывать только один исполнитель. Это реализовано при помощи *sync.Mutex.

    b) Количество одновременно вычисляющихся хешей ограничено константой N (в данной программе ее значение равно 10), что осуществлено с помощью экспериментального пакета "golang.org/x/sync/semaphore".

6. Путь ```/rest/self```

GET /rest/self/find/$str

Функционал реализован хендлером GetIdentifiers. Но что-то пошло не так :)))

Завершение программы:

```
docker-compose down
```

```
docker rmi -f $(docker images -aq)
```