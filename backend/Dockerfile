FROM golang:alpine AS builder

RUN mkdir /src
ADD . /src
WORKDIR /src
RUN go mod tidy -v
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/main.go


FROM alpine:latest AS production
RUN apk add --no-cache tzdata ca-certificates
# Set timezone ke Asia/Jakarta
ENV TZ=Asia/Jakarta

WORKDIR /app
COPY --from=builder /src/api .
COPY --from=builder /src/templates ./templates
COPY --from=builder /src/docs ./docs

# Create uploads directory
RUN mkdir -p /app/uploads/products /app/uploads/assets

EXPOSE 8080
CMD ["./api"]