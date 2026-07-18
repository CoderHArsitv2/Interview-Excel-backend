FROM golang:1.23.2-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /interviewexcel-backend ./main.go

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /interviewexcel-backend /app/interviewexcel-backend
COPY --from=builder /app/config/*.yaml /app/config/

ARG BUILD_ENV=staging
LABEL environment=${BUILD_ENV}

ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/interviewexcel-backend"]
CMD ["serve"]
