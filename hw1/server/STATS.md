# Статистика тестирования web-сервера
Отчет по нагрузочному тестированию web-сервара, использующего nginx
## Без кэширования и без балансировки

Server Hostname:        localhost

Server Port:            3333

Document Path:          /

Document Length:        258891 bytes

Concurrency Level:      100

Time taken for tests:   37.062 seconds

Complete requests:      10000

Failed requests:        0

Total transferred:      2589930000 bytes

HTML transferred:       2588910000 bytes

Requests per second:    269.82 [#/sec] (mean)

Time per request:       370.616 [ms] (mean)

Time per request:       3.706 [ms] (mean, across all concurrent requests)

Transfer rate:          68243.85 [Kbytes/sec] received

Connection Times (ms)

|            | min | mean | [+/-sd] | median |  max |
|------------|:---:|:----:|:-------:|:------:|-----:|
| Connect    |  0  |  0   |   0.2   |   0    |    4 |
| Processing | 12  | 369  |  374.7  |  243   | 3211 |
| Waiting    | 12  | 366  |  373.3  |  240   | 3211 |
| Total      | 12  | 369  |  374.7  |  243   | 3212 |

## Только кэширование

Server Software:        nginx/1.18.0

Server Hostname:        localhost

Server Port:            90

Document Path:          /

Document Length:        258891 bytes

Concurrency Level:      100

Time taken for tests:   1.600 seconds

Complete requests:      10000

Failed requests:        0

Total transferred:      2590430000 bytes

HTML transferred:       2588910000 bytes

Requests per second:    6250.66 [#/sec] (mean)

Time per request:       15.998 [ms] (mean)

Time per request:       0.160 [ms] (mean, across all concurrent requests)

Transfer rate:          1581241.00 [Kbytes/sec] received

Connection Times (ms)

|            | min | mean | [+/-sd] | median | max |
|------------|:---:|:----:|:-------:|:------:|----:|
| Connect    |  0  |  1   |   0.2   |   1    |   3 |
| Processing |  2  |  15  |   2.1   |   15   |  28 |
| Waiting    |  0  |  1   |   0.7   |   1    |  11 |
| Total      |  5  |  16  |   2.2   |   16   |  29 |

## Только балансировка нагрузки (два идентичных сервера)

Server Software:        nginx/1.18.0

Server Hostname:        localhost

Server Port:            90

Document Path:          /

Document Length:        258891 bytes

Concurrency Level:      100

Time taken for tests:   32.976 seconds

Complete requests:      10000

Failed requests:        0

Total transferred:      2590430000 bytes

HTML transferred:       2588910000 bytes

Requests per second:    303.25 [#/sec] (mean)

Time per request:       329.760 [ms] (mean)

Time per request:       3.298 [ms] (mean, across all concurrent requests)

Transfer rate:          76713.96 [Kbytes/sec] received

Connection Times (ms)

|            | min | mean | [+/-sd] | median |  max |
|------------|:---:|:----:|:-------:|:------:|-----:|
| Connect    |  0  |  0   |   0.3   |   0    |    5 |
| Processing | 14  | 328  |  334.2  |  211   | 2552 |
| Waiting    | 13  | 324  |  332.5  |  207   | 2551 |
| Total      | 14  | 328  |  334.2  |  211   | 2552 |

## Все вместе

Server Software:        nginx/1.18.0

Server Hostname:        localhost

Server Port:            90

Document Path:          /

Document Length:        258891 bytes

Concurrency Level:      100

Time taken for tests:   1.581 seconds

Complete requests:      10000

Failed requests:        0

Total transferred:      2590430000 bytes

HTML transferred:       2588910000 bytes

Requests per second:    6326.01 [#/sec] (mean)

Time per request:       15.808 [ms] (mean)

Time per request:       0.158 [ms] (mean, across all concurrent requests)

Transfer rate:          1600300.61 [Kbytes/sec] received

Connection Times (ms)

|            | min | mean | [+/-sd] | median | max |
|------------|:---:|:----:|:-------:|:------:|----:|
| Connect    |  0  |  1   |   0.3   |   1    |   5 |
| Processing |  4  |  15  |   1.7   |   15   |  25 |
 | Waiting    |  0  |  1   |   0.5   |   1    |   9 |
 | Total      |  6  |  16  |   1.7   |   15   |  27 |
