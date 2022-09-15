# The `cron` package

[![cron](https://github.com/alex-schneider/cron/actions/workflows/go.yaml/badge.svg)](https://github.com/alex-schneider/cron/actions/workflows/go.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex-schneider/cron)](https://goreportcard.com/report/github.com/alex-schneider/cron)

The `cron` package implements a cron job scheduler with a lot of non-standard extensions.
It allowes to run jobs periodically at fixed times, dates, frequencies and intervals.

## Format

The cron `expression` is a string with 5, 6 or 7 fields separated by any number of
whitespace like `SP` (`0x20`) or `HT` (`0x09`) from the ASCII table. Additionally,
this `cron` package implements and supports a lot of [Non-Standard Macros](#non-standard-macros)
as listed below.

### Fields

The standard crontab supports only 5 fields and cannot be scaled to seconds or years.
The current implementation completes the fields as following:

* If only 5 fields are present, the `seconds` field with value `0` is added to the
  begin and the `year` field with value `*` is added to the end of the fields list.
* If only 6 fields are present, the `year` field with value `*` is added to the
  end of the fields list.

| Field name | Required | Description               | Allowed values      | Allowed special characters          |
| :--------- | :------: | :------------------------ | :------------------ | :---------------------------------- |
| `seconds`  | &#10005; | Represents seconds.       | `0-59`              | `,` `-` `*` `/` `.` `R`             |
| `minutes`  | &#10003; | Represents minutes.       | `0-59`              | `,` `-` `*` `/` `.` `R`             |
| `hours`    | &#10003; | Represents hours.         | `0-23`              | `,` `-` `*` `/` `.` `R`             |
| `dom`      | &#10003; | Represents days-of-month. | `1-31`              | `,` `-` `*` `/` `.` `R` `?` `L` `W` |
| `month`    | &#10003; | Represents months.        | `1-12` or `JAN-DEC` | `,` `-` `*` `/` `.` `R`             |
| `dow`      | &#10003; | Represents days-of-week.  | `0-7` or `SUN-SAT`  | `,` `-` `*` `/` `.` `R` `?` `L` `#` |
| `year`     | &#10005; | Represents years.         | `1970-2099`         | `,` `-` `*` `/` `.` `R`             |

---

> **Note:** Both, `0` and `7` value in the `dow` (days-of-week) field is interpreted
 as `SUN` (Sunday).

---

> **Note:** The names in `month` and `dow` (days-of-week) fields and the special
 characters `R`, `L` and `W` are case insensitive. For example, `FRI` is the same
 as `Fri` or `fri`.

---

### Special Characters

| Special Character | Description |
| :---------------: | :---------- |
| `,`               | Commas are used to separate values in a field. For example, using `MON,FRI,SUN` in the `dow` (days-of-week) field means `Monday, Friday and Sunday`. |
| `-`               | Hyphen defines ranges. For example, `JAN-MAR` in the `month` field means `Januar, Februar and March`. The current implementation supports ranges with "range-overflows" for all expression fields excepting the `year` field. For example, `FRI-MON` in the `dow` (days-of-week) field means `Friday, Saturday, Sunday and Monday`. A mix of names and numeric values in `month` and `dow` (days-of-week) fields is supported, too. For example, `JAN-MAR` is the same as `JAN-3`. |
| `*`               | Asterisk is used to select all possible values within a field. For example, `*` in the `month` field means `daily` or `every day`. |
| `/`               | Slash can be used to specify frequencies. For example, `*/10` in the `seconds` field means `every 10 seconds`. And `10/15` in the `minutes` field means `the minutes 10, 25, 40 and 55`. |
| `.`               | Dot can be used to specify the current date or time value on the startup. For example, `0 . . * * * *` would be updated to `0 9 15 * * * *` if the cron is started-up at 09:15. |
| `?`               | Question mark is used for leaving either, `dom` (day-of-month) or `dow` (day-of-week) blank. For example, `0 0 0 15 * ? *` would trigger the cronjob at `15th` of every month regardless of what day-of-week it is. |
| `R`               | `R` stands for `random`. `R` can be combined with ranges, e.g. `10-30/R` in the `minutes` field. Once generated during parsing, the random number remains constant for current field. If used in the `dom` (day-of-month) field without ranges, the possible values are limited to the range `1-28`. To be able to use the total range set `1-31/R` to the `dom` (day-of-month) field. |
| `L`               | `L` stands for `last`. When this character is used in the `dom` (day-of-month) field, it specifies the last day of the month. For example, `31 January` or `29 February` in a leap year. In the `dow` (day-of-week) field, it specifies the last day of the week and simply means the `SAT` or `6`. When this character is used in the `dow` (day-of-week) field and is prefixed with a number, it means `the last X day of the month`. For example, `1L` means the `last Monday of the month`. `MONL` is the same as `1L`. |
| `W`               | `W` stands for `weekday` (Monday-Friday). `W` is used to specify the business day nearest the given day in the given month. It never jumps over the boundary of the month's days. For example, if `1W` is a Saturday, the cronjob would trigger at Monday, the 3rd. The `L` and `W` special characters can also be combined in the `dom` (day-of-month) field as `LW`, which means `last weekday of the month`. |
| `#`               | Hash allows to specifying constructs such as `the second Friday` of a given month. For example, `5#3` in the `dow` (day-of-week) field means `the third Friday of every month`. The value before the `#` has the range `0-7` or `SUN-SAT`. The value after the `#` has the range `1-5`. |

---

> **Note:** Be careful with the use of `?` and `*` special characters in the `dom`
 (day-of-month) or `dow` (day-of-week) fields.

---

> **Note:** Avoid cronjobs at `daylight saving times`. Cronjobs can be skiped or
 repeated depending on whether the time moves back or jumps forward.

---

> **Note:** It is not allowed to combine the `?` with other values within a field.
 This special character can only be used exclusively within a field.

---

> **Note:** It is not allowed to define the `?` special character in both, `dom`
 (days-of-month) and `dow` (days-of-week) fields at the same time.

---

> **Note:** If both, the `dom` (day-of-month) and `dow` (day-of-week) fields contain
 any values excepting the `?`, the day is calculated from the best possible value
 of the both fields.

---

### Non-Standard Macros

| Macro           | Description                                                 | Equivalent expression |
| :-------------- | :---------------------------------------------------------- | :-------------------: |
| `@yearly`       | Run once a year at midnight of 1 January.                   | `0 0 0 1 1 * *`       |
| `@annually`     | The same as `@yearly`.                                      | `0 0 0 1 1 * *`       |
| `@monthly`      | Run once a month at midnight of the first day of the month. | `0 0 0 1 * * *`       |
| `@weekly`       | Run once a week at midnight on Sunday.                      | `0 0 0 * * 0 *`       |
| `@daily`        | Run once a day at midnight.                                 | `0 0 0 * * * *`       |
| `@midnight`     | The same as `@daily`.                                       | `0 0 0 * * * *`       |
| `@hourly`       | Run once an hour at the beginning of the hour.              | `0 0 * * * * *`       |
| `@minutely`     | Run once a minute at the beginning of the minute.           | `0 * * * * * *`       |
| `@every_minute` | The same as `@minutely`.                                    | `0 * * * * * *`       |
| `@secondly`     | Run secondly.                                               | `* * * * * * *`       |
| `@every_second` | The same as `@secondly`.                                    | `* * * * * * *`       |
| `@reboot`       | Run once at startup.                                        | &#10005;              |

## Examples

| Expression           | Description                                                |
| :------------------- | :--------------------------------------------------------- |
| `0 11 11 11 11 ? *`  | Run every November 11th at 11:11am.                        |
| `59 59 23 31 12 ? *` | Run at every turn of the year.                             |
| `0 0 0 ? * LW *`     | Run at every last business day on the month.               |
| `0 0 0 ? * 5L *`     | Run at every last Friday on the month.                     |
| `0 0 0 29 2 ? *`     | Run every February 29th on every leap year.                |
| `0 0 R * * * *`      | Run once a day at a random hour.                           |
| `0 0 2-6/R * * * *`  | Run once a day at a random hour between 2:00am and 6:00am. |
