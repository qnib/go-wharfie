FROM qnib/alplain-golang:1.9.2 AS build

WORKDIR /usr/local/src/github.com/qnib/go-wharfie/
COPY . ./
RUN go install

FROM qnib/alplain-init

COPY --from=build /usr/local/bin/go-wharfie /usr/bin/
COPY start.sh /opt/go-wharfie/bin/start.sh
CMD ["/opt/go-wharfie/bin/start.sh"]
