# Сначала мы должны получить копию gotlang, мы будем использовать alpine, потому что это компактная версия golang
FROM golang:alpine as builder
# Вы можете добавить метку в свой файл docker и записать в нее информацию о владельце
LABEL maintainer="amoCRM"
#Как вы знаете, когда вы используете go, вы должны иметь возможность загружать и устанавливать библиотеки с github, вот почему мы можем установить git внутри нашего контейнера
RUN apk update && apk add --no-cache git
#теперь мы можем установить наш рабочий директорию внутри нашего контейнера
WORKDIR /app
#Мы должны скопировать go.mod и go.some, чтобы иметь возможность загружать все зависимости
COPY ./go.mod  go.sum ./
# Чтобы загрузить все зависимости, когда мы создаем контейнер из нашего образа, мы можем задать команду RUN, которая сделает это за нас
RUN go mod download
# Скопируйте исходный код из текущего директории в рабочий каталог внутри контейнера
COPY . .
# Создать приложение Go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .
# Начните новый этап с нуля
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
#Скопировать предварительно созданный двоичный файл с предыдущего этапа. Обратите внимание, что мы также скопировали файл .env
COPY --from=builder /app/main .
COPY --from=builder /app/.env .    
# Expose port 8080 to the outside world
EXPOSE 8080
#Команда для запуска исполняемого файла
CMD ["/app/main"]
