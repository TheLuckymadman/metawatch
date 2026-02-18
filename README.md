# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**
File: server
Build ID: f586d2cb0304a66b1ec3515266264e09bc26e28c
Type: inuse_space
Time: 2026-02-01 16:19:22 MSK
Showing nodes accounting for -306.33kB, 6.79% of 4508.75kB total
      flat  flat%   sum%        cum   cum%
   -2052kB 45.51% 45.51%    -2052kB 45.51%  runtime.allocm
  650.62kB 14.43% 31.08%  1233.63kB 27.36%  compress/flate.(*compressor).init
  583.01kB 12.93% 18.15%   583.01kB 12.93%  compress/flate.newDeflateFast (inline)
  512.04kB 11.36%  6.79%   512.04kB 11.36%  fmt.Sprintln
         0     0%  6.79%  1233.63kB 27.36%  compress/flate.NewWriter (inline)
         0     0%  6.79%  1233.63kB 27.36%  compress/gzip.(*Writer).Write
         0     0%  6.79%  1233.63kB 27.36%  github.com/TheLuckymadman/metawatch/internal/handler.(*ResponseWriterCompressor).Write
         0     0%  6.79%  1233.63kB 27.36%  github.com/TheLuckymadman/metawatch/internal/handler.CompressWrapper.func1
         0     0%  6.79%  1745.67kB 38.72%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0%  6.79%  1745.67kB 38.72%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0%  6.79%   512.04kB 11.36%  go.uber.org/zap.(*SugaredLogger).Infoln
         0     0%  6.79%   512.04kB 11.36%  go.uber.org/zap.(*SugaredLogger).logln
         0     0%  6.79%   512.04kB 11.36%  go.uber.org/zap.getMessageln (inline)
         0     0%  6.79%  1233.63kB 27.36%  main.run.HashWrapper.func10.1
         0     0%  6.79%  1233.63kB 27.36%  main.run.JSONSetterHandler.func8
         0     0%  6.79%  1745.67kB 38.72%  main.run.LoggerWrapper.func9.1
         0     0%  6.79%  1745.67kB 38.72%  net/http.(*conn).serve
         0     0%  6.79%  1745.67kB 38.72%  net/http.HandlerFunc.ServeHTTP
         0     0%  6.79%  1745.67kB 38.72%  net/http.serverHandler.ServeHTTP
         0     0%  6.79%    -1539kB 34.13%  runtime.mcall
         0     0%  6.79%     -513kB 11.38%  runtime.morestack
         0     0%  6.79%    -2052kB 45.51%  runtime.newm
         0     0%  6.79%     -513kB 11.38%  runtime.newstack
         0     0%  6.79%    -1539kB 34.13%  runtime.park_m
         0     0%  6.79%     -513kB 11.38%  runtime.preemptPark
         0     0%  6.79%    -2052kB 45.51%  runtime.resetspinning
         0     0%  6.79%    -2052kB 45.51%  runtime.schedule
         0     0%  6.79%    -2052kB 45.51%  runtime.startm
         0     0%  6.79%    -2052kB 45.51%  runtime.wakep
