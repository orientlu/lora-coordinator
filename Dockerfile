FROM golang:1.11-alpine AS development

ENV PROJECT_PATH=/lora-coordinator
ENV PATH=$PATH:$PROJECT_PATH/build
ENV CGO_ENABLED=0
ENV GO_EXTRA_BUILD_ARGS="-a -installsuffix cgo"

RUN apk add --no-cache ca-certificates make git bash\
    && mkdir -p $PROJECT_PATH
WORKDIR $PROJECT_PATH

COPY ./go.mod  $PROJECT_PATH
RUN go mod download

COPY . .
RUN make dev-requirements
RUN make

# -----
FROM alpine:latest AS production
WORKDIR /root/
RUN apk --no-cache add ca-certificates\
COPY --from=development /lora-coordinator/build/lora-coordinator .
ENTRYPOINT ["./lora-coordinator"]
