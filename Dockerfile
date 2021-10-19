FROM golang:alpine
ADD . /app
WORKDIR /app
RUN go get ; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
CMD ./Crisp_Telegram_bot

