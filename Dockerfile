FROM debian:stable-slim

WORKDIR app

COPY safina-society-search safina-society-search

COPY public/ public/

CMD ["./safina-society-search"]
