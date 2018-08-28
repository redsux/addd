FROM scratch
COPY --from=addd-bld:latest /go/bin/addd /addd
EXPOSE 53/udp
ENTRYPOINT ["/addd"]
CMD ["-port=53"]