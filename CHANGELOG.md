# go-alphasort changelog

## 0.0.3

- embedded struct fields go before regular struct fields

## 0.0.2

- constructors are all functions that start with "New" + uppercase character and whose first return value contains an instance of the type

## 0.0.1

- sorts freestanding functions alphabetically (public first)
  - except in files called "main.go" because those are script files ordered based on call flows
- sorts struct fields alphabetically
- releases for Linux, macOS, and Windows for ARM64 and Intel64
- `alphasort version` displays the version of alphasort
