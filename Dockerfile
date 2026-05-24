# syntax=docker/dockerfile:1

FROM golang:1.26

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.19.1

RUN env CGO_ENABLED=0 go build -v -x -o /daily-planet .

CMD ["sh", "-c", "migrate -path /app/db/migrations -database 'sqlite3:///data/daily_planet.db' up && /daily-planet"]
