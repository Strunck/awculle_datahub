FROM cosmtrek/air

RUN go install github.com/a-h/templ/cmd/templ@latest

ENTRYPOINT ["/go/bin/air"]