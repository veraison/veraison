
FROM alpine:latest AS builder

RUN apk update && apk upgrade && apk add make go gcc g++ bash

WORKDIR /opt/veraison

COPY . .

RUN  make clean && make

FROM alpine:latest

WORKDIR /opt/veraison

COPY --from=builder /opt/veraison/frontend/verifier /opt/veraison/verifier
COPY --from=builder /opt/veraison/plugins/bin/ /opt/veraison/plugins/
COPY --from=builder /opt/veraison/frontend/test/db/ /opt/veraison/db/

EXPOSE 8080

CMD ["./verifier", "-p", "plugins/", "-d", "db/"]
