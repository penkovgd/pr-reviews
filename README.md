# **Сервис назначения ревьюеров для Pull Request’ов**

[Условие](./Backend-trainee-assignment-autumn-2025.md)

Написал микросервис на go, по гексагональной архитектуре. Получилось немного громоздко, но гибко и расширяемо).
БД - postgresql, Есть конфигурация через config.yaml или переменные окружения. Миграции применяются при запуске приложения

## Запуск

1. при необходимости установить инструменты разработки (bombardier, golangci-lint):

> make tools

2. Запустить все стразу (сервис с миграциями, интеграционные и нагрузочные тесты)

> make test

Ожидаемые результаты:

- Интеграционные тесты:

```bash
=== RUN   TestTeamCreate_Get
--- PASS: TestTeamCreate_Get (0.02s)
=== RUN   TestUserSetIsActive
--- PASS: TestUserSetIsActive (0.02s)
=== RUN   TestPRCreate_AssignsUpToTwoActiveReviewers
--- PASS: TestPRCreate_AssignsUpToTwoActiveReviewers (0.03s)
=== RUN   TestPRCreate_NoCandidates
--- PASS: TestPRCreate_NoCandidates (0.03s)
=== RUN   TestPRMerge_Idempotent
--- PASS: TestPRMerge_Idempotent (0.03s)
=== RUN   TestPRReassign_Success
--- PASS: TestPRReassign_Success (0.03s)
=== RUN   TestPRReassign_FailOnMerged
--- PASS: TestPRReassign_FailOnMerged (0.03s)
=== RUN   TestPRReassign_FailNoCandidate
--- PASS: TestPRReassign_FailNoCandidate (0.02s)
=== RUN   TestUsersGetReview
--- PASS: TestUsersGetReview (0.02s)
PASS
ok      tests   1.263s
```

- Простой нагрузочный тест (создаем 1 команду и читаем ее через 20 одновременный подключений)

```bash
curl -X POST "http://localhost:8080/team/add" \
-H "Content-Type: application/json" \
-d '{"team_name":"test-team","members":[{"user_id":"user1","username":"Test User","is_active":true}]}'
{"team":{"team_name":"test-team","members":[{"user_id":"user1","username":"Test User","is_active":true}]}}
bombardier -c 20 -d 15s -l "http://localhost:8080/team/get?team_name=test-team"
Bombarding http://localhost:8080/team/get?team_name=test-team for 15s using 20 connection(s)
[=====================================================================================================================] 15s
Done!
Statistics        Avg      Stdev        Max
  Reqs/sec      1692.60     293.96    2515.93
  Latency       11.80ms     8.71ms    73.64ms
  Latency Distribution
     50%     5.49ms
     75%    21.30ms
     90%    26.45ms
     95%    29.82ms
     99%    39.99ms
  HTTP codes:
    1xx - 0, 2xx - 25399, 3xx - 0, 4xx - 0, 5xx - 0
    others - 0
  Throughput:   489.01KB/s
```

Как видно:

- RPS ~ 1600
- SLI времени ответа (99 перцентиль) ~ 40мс
- SLI успешности = 100% (тестили недолго, но все же)

3. При желании можно проверить линтером:

> make lint

## Доп. задания

1. Сделал нагрузочное тестирование
2. Написал интеграционные тесты
3. Описал линтер [(конфигрурация линтера)](./.golangci.yaml)
4. Написал простой эндпоинт статистики, который показывает кол-во назначений pr-ов для каждого пользователя:

```bash
$ curl localhost:8080/stats/user-assignments
{"user_assignments":{"author":0,"author-a":0,"author-b":0,"author-c":0,"author-d":0,"author-f":0,"cand":1,"candidate":1,"oldrev":2,"r1":2,"r10":2,"r2":2,"rev":4,"rev-f":2,"u1":0,"u2":0,"user-activate":0}}
```

## Вопросы/Проблемы с которыми столкнулся

1. При переназначении ревьювера может быть такое, что кандидатов не найдется. В таком случае можно было бы придумать новую ошибку, но я решил никого не назначать. Потом увидел в сваггере, что там такой пример оказывается был, надо было лишь выбрать его :). В итоге возвращается ошибка NO_CANDIDATE
